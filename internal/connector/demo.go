package connector

import (
	"context"
	"fmt"
	"time"

	"github.com/noahxsu777/Gooo/internal/events"
)

type DemoConnector struct {
	Interval      time.Duration
	DropEvery     int
	LatencySample time.Duration
}

func (d DemoConnector) Name() string { return "demo" }

func (d DemoConnector) Connect(ctx context.Context, channel string) (<-chan events.Message, <-chan error) {
	out := make(chan events.Message)
	errs := make(chan error, 1)
	if d.Interval <= 0 {
		d.Interval = 1200 * time.Millisecond
	}
	go func() {
		defer close(out)
		defer close(errs)
		tk := time.NewTicker(d.Interval)
		defer tk.Stop()
		count := 0
		for {
			select {
			case <-ctx.Done():
				return
			case t := <-tk.C:
				count++
				if d.DropEvery > 0 && count%d.DropEvery == 0 {
					errs <- fmt.Errorf("simulación de caída de conexión demo")
					return
				}
				msg := events.Message{
					ID:        fmt.Sprintf("%s-%d", channel, count),
					Type:      "chat",
					User:      fmt.Sprintf("usuario%d", count%5),
					Text:      fmt.Sprintf("Mensaje demo #%d", count),
					Priority:  events.PriorityNormal,
					Source:    "demo",
					Timestamp: t,
				}
				select {
				case out <- msg:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return out, errs
}
