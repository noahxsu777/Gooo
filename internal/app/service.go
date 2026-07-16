package app

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/noahxsu777/Gooo/internal/backoff"
	"github.com/noahxsu777/Gooo/internal/config"
	"github.com/noahxsu777/Gooo/internal/connector"
	"github.com/noahxsu777/Gooo/internal/events"
	"github.com/noahxsu777/Gooo/internal/metrics"
	"github.com/noahxsu777/Gooo/internal/queue"
	"github.com/noahxsu777/Gooo/internal/tts"
	"github.com/noahxsu777/Gooo/internal/webhook"
)

type Service struct {
	cfg     config.Config
	logger  *slog.Logger
	broker  *Broker
	metrics *metrics.Metrics
	queue   *queue.PriorityQueue
	deduper *events.Deduper
	filter  *tts.Filters
	limiter *webhook.RateLimiter

	backoff   backoff.Exponential
	connector connector.Connector

	mu        sync.RWMutex
	state     events.State
	ready     bool
	cancelRun context.CancelFunc
}

type webhookPayload struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	User     string `json:"user"`
	Text     string `json:"text"`
	Priority int    `json:"priority"`
}

func NewService(cfg config.Config, logger *slog.Logger) *Service {
	s := &Service{
		cfg:     cfg,
		logger:  logger,
		broker:  NewBroker(),
		metrics: &metrics.Metrics{},
		queue:   queue.New(cfg.QueueSize),
		deduper: events.NewDeduper(5 * time.Minute),
		filter: &tts.Filters{
			MaxLength:      300,
			IgnoreCommands: true,
			BlockedUsers:   map[string]struct{}{},
			BlockedTypes:   map[string]struct{}{"gift": {}},
			SpamWindow:     1,
		},
		backoff: backoff.Exponential{
			Base:    cfg.ReconnectBase,
			Max:     cfg.ReconnectMax,
			Jitter:  0.35,
			Retries: cfg.ReconnectMaxAttempts,
		},
		state: events.State{
			Status:    events.StatusDisconnected,
			Channel:   cfg.Channel,
			UpdatedAt: time.Now(),
		},
		limiter: webhook.NewRateLimiter(30, time.Minute),
	}
	if cfg.Mode == "demo" {
		s.connector = connector.DemoConnector{Interval: cfg.DemoInterval, DropEvery: cfg.DemoDropEvery}
	}
	return s
}

func (s *Service) Start(ctx context.Context) {
	runCtx, cancel := context.WithCancel(ctx)
	s.cancelRun = cancel
	go s.run(runCtx)
}

func (s *Service) Stop() {
	if s.cancelRun != nil {
		s.cancelRun()
	}
}

func (s *Service) run(ctx context.Context) {
	attempt := 0
	for {
		if !s.backoff.CanRetry(attempt) {
			s.updateState(func(st *events.State) {
				st.Status = events.StatusDisconnected
				st.LastError = "límite de reconexiones alcanzado"
			})
			return
		}
		if err := s.connectOnce(ctx); err != nil && !errors.Is(err, context.Canceled) {
			attempt++
			s.metrics.IncReconnects()
			wait := s.backoff.Duration(attempt)
			s.updateState(func(st *events.State) {
				st.Status = events.StatusReconnecting
				st.LastError = err.Error()
				st.Reconnects = attempt
			})
			s.logger.Warn("connector disconnected, retrying", "attempt", attempt, "wait", wait, "error", err)
			select {
			case <-ctx.Done():
				return
			case <-time.After(wait):
				continue
			}
		}
		attempt = 0
		select {
		case <-ctx.Done():
			return
		default:
		}
	}
}

func (s *Service) connectOnce(ctx context.Context) error {
	if s.connector == nil {
		s.updateState(func(st *events.State) {
			st.Status = events.StatusDisconnected
			st.LastError = "no hay conector externo configurado; usa modo demo o configura adaptador"
		})
		<-ctx.Done()
		return ctx.Err()
	}
	s.updateState(func(st *events.State) {
		st.Status = events.StatusConnecting
		st.LastError = ""
	})
	conCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	msgCh, errCh := s.connector.Connect(conCtx, s.cfg.Channel)
	s.setReady(true)
	heartbeat := time.NewTimer(s.cfg.HeartbeatTimeout)
	defer heartbeat.Stop()
	s.updateState(func(st *events.State) { st.Status = events.StatusLive })
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-errCh:
			if err == nil {
				return errors.New("conector finalizó sin detalle")
			}
			return err
		case <-heartbeat.C:
			return errors.New("timeout de heartbeat")
		case msg, ok := <-msgCh:
			if !ok {
				return errors.New("stream cerrado")
			}
			if !heartbeat.Stop() {
				select {
				case <-heartbeat.C:
				default:
				}
			}
			heartbeat.Reset(s.cfg.HeartbeatTimeout)
			s.handleMessage(msg)
		}
	}
}

func (s *Service) handleMessage(msg events.Message) {
	if s.deduper.IsDuplicate(msg.ID) {
		return
	}
	if !s.filter.ShouldSpeak(msg) {
		return
	}
	s.metrics.IncEventsIn()
	if ok := s.queue.Enqueue(msg); !ok {
		s.metrics.AddDropped(1)
	}
	s.updateState(func(st *events.State) {
		st.LastEvent = &msg
		st.QueueSize = s.queue.Len()
		st.DroppedMessages = s.queue.Dropped() + s.metrics.Dropped()
		st.LatencyMS = time.Since(msg.Timestamp).Milliseconds()
	})
	s.broker.Publish("chat", msg)
	s.broker.Publish("state", BuildStateEvent(s.State()))
}

func (s *Service) DequeueMessage() (events.Message, bool) {
	msg, ok := s.queue.Dequeue()
	if ok {
		s.metrics.IncEventsOut()
		s.updateState(func(st *events.State) { st.QueueSize = s.queue.Len() })
	}
	return msg, ok
}

func (s *Service) ServeEvents(w http.ResponseWriter, r *http.Request) {
	if origin := r.Header.Get("Origin"); s.cfg.AllowedOrigin != "*" && origin != "" && origin != s.cfg.AllowedOrigin {
		http.Error(w, "origen no permitido", http.StatusForbidden)
		return
	}
	s.broker.ServeHTTP(w, r)
}

func (s *Service) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "método no permitido", http.StatusMethodNotAllowed)
		return
	}
	if !s.limiter.Allow(r.RemoteAddr) {
		s.metrics.IncWebhookReject()
		http.Error(w, "rate limit excedido", http.StatusTooManyRequests)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, s.cfg.MaxPayloadBytes)
	body, err := readAll(r)
	if err != nil {
		s.metrics.IncWebhookReject()
		http.Error(w, "payload inválido", http.StatusBadRequest)
		return
	}
	sig := r.Header.Get(s.cfg.TikToolsWebhookHeader)
	if !webhook.Validate(s.cfg.WebhookSecret, sig, body) {
		s.metrics.IncWebhookReject()
		http.Error(w, "firma inválida", http.StatusUnauthorized)
		return
	}
	var p webhookPayload
	if err := json.Unmarshal(body, &p); err != nil || p.ID == "" || p.Text == "" {
		s.metrics.IncWebhookReject()
		http.Error(w, "json inválido", http.StatusBadRequest)
		return
	}
	msg := events.Message{
		ID:        p.ID,
		Type:      p.Type,
		User:      p.User,
		Text:      p.Text,
		Priority:  events.Priority(p.Priority),
		Source:    "tiktools-webhook",
		Timestamp: time.Now(),
	}
	s.handleMessage(msg)
	w.WriteHeader(http.StatusAccepted)
	_, _ = w.Write([]byte(`{"ok":true}`))
}

func readAll(r *http.Request) ([]byte, error) {
	defer r.Body.Close()
	return io.ReadAll(r.Body)
}

func (s *Service) State() events.State {
	s.mu.RLock()
	defer s.mu.RUnlock()
	copy := s.state
	return copy
}

func (s *Service) updateState(fn func(*events.State)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	fn(&s.state)
	s.state.UpdatedAt = time.Now()
}

func (s *Service) setReady(v bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ready = v
}

func (s *Service) Healthz(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func (s *Service) Readyz(w http.ResponseWriter, _ *http.Request) {
	s.mu.RLock()
	ready := s.ready
	s.mu.RUnlock()
	if !ready {
		http.Error(w, "not ready", http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ready"))
}

func (s *Service) Metrics(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	_, _ = w.Write([]byte(s.metrics.PrometheusFormat()))
}

type controlRequest struct {
	Action string `json:"action"`
}

func (s *Service) HandleControl(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "método no permitido", http.StatusMethodNotAllowed)
		return
	}
	var c controlRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 2048)).Decode(&c); err != nil {
		http.Error(w, "json inválido", http.StatusBadRequest)
		return
	}
	switch c.Action {
	case "next":
		msg, ok := s.DequeueMessage()
		if !ok {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(msg)
	case "flush":
		s.queue.Clear()
		s.updateState(func(st *events.State) { st.QueueSize = 0 })
		w.WriteHeader(http.StatusNoContent)
	default:
		http.Error(w, "acción no soportada", http.StatusBadRequest)
	}
}

func (s *Service) HandleState(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(s.State())
}
