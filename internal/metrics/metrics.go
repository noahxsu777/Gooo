package metrics

import (
	"fmt"
	"sync/atomic"
)

type Metrics struct {
	eventsIn      atomic.Int64
	eventsOut     atomic.Int64
	dropped       atomic.Int64
	reconnects    atomic.Int64
	webhookReject atomic.Int64
}

func (m *Metrics) IncEventsIn()       { m.eventsIn.Add(1) }
func (m *Metrics) IncEventsOut()      { m.eventsOut.Add(1) }
func (m *Metrics) AddDropped(n int64) { m.dropped.Add(n) }
func (m *Metrics) IncReconnects()     { m.reconnects.Add(1) }
func (m *Metrics) IncWebhookReject()  { m.webhookReject.Add(1) }
func (m *Metrics) Dropped() int64     { return m.dropped.Load() }
func (m *Metrics) Reconnects() int64  { return m.reconnects.Load() }
func (m *Metrics) PrometheusFormat() string {
	return fmt.Sprintf(
		"events_in_total %d\nevents_out_total %d\ndropped_total %d\nreconnect_total %d\nwebhook_reject_total %d\n",
		m.eventsIn.Load(),
		m.eventsOut.Load(),
		m.dropped.Load(),
		m.reconnects.Load(),
		m.webhookReject.Load(),
	)
}
