package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/kmpm/enp/internal/eddn"
	"github.com/kmpm/enp/internal/message"
	"github.com/kmpm/enp/internal/sink"
	"github.com/kmpm/enp/public/models"

	"github.com/nats-io/jsm.go/natscontext"
	"github.com/nats-io/nats.go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Define Prometheus metrics
var (
	messageCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "enp_message_counter",
		Help: "The total number of messages",
	}, []string{"status"})

	messageDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "enp_message_duration",
		Help:    "The duration of messages",
		Buckets: []float64{0.0005, .001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5},
	}, []string{"status"})
)

type RunCmd struct {
}

func (cmd *RunCmd) Run() error {
	stop := waitfor()
	wg := &sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	validator, err := message.NewValidator()
	if err != nil {
		return err
	}

	nc, err := connect()
	if err != nil {
		return err
	}

	ch := make(chan []byte, 5)
	wg.Add(1)
	slog.Info("starting eddn subscriber")
	go eddn.Subscribe(ctx, wg, ch)
	run := true
	for run {
		select {
		case rawMsg := <-ch:
			err = processMessage(nc, validator, rawMsg)
			if err != nil {
				slog.Error("error processing message", "error", err)
			}
		case <-stop:
			run = false
			cancel()
		}
	}
	wg.Wait()
	return nil
}

func processMessage(nc *nats.Conn, validator *message.Validator, rawMsg []byte) error {
	start := time.Now()
	status := "published"
	schema := "unknown"
	defer func() {
		messageDuration.WithLabelValues(status).Observe(float64(time.Since(start).Milliseconds()))
		if time.Since(start) > 1*time.Second {
			slog.Warn("slow message", "duration", time.Since(start), "status", status, "schema", schema)
		}
	}()
	deflated, err := message.Deflate(rawMsg)
	if err != nil {
		status = "deflate_error"
		return fmt.Errorf("error deflating message: %w", err)
	}
	var decoded models.EDDN
	err = json.Unmarshal(deflated, &decoded)
	if err != nil {
		status = "decode_error"
		return fmt.Errorf("error decoding message: %w", err)
	}
	schema = decoded.SchemaRef

	err = validator.Validate(decoded.SchemaRef, bytes.NewReader(deflated))
	if err != nil {
		status = "validation_error"
		// errf := os.WriteFile("var/invalid/"+decoded.Header.UploaderID+".dat", rawMsg, 0644)
		// if errf != nil {
		// 	slog.Error("error writing invalid message", "error", errf)
		// }
		return fmt.Errorf("error validating message: %w", err)
	}
	// slog.Debug("received message", "schema", decoded.SchemaRef, "software", decoded.Header.SoftwareName)
	err = sink.Publish(nc, &decoded, rawMsg)
	if err != nil {
		status = "publish_error"
		return fmt.Errorf("error publishing message: %w", err)
	}
	return nil
}

func connect() (*nats.Conn, error) {
	nc, err := natscontext.Connect("nats_development", nil)
	nc.SetClosedHandler(func(_ *nats.Conn) {
		slog.Error("connection closed")
	})
	nc.SetErrorHandler(func(_ *nats.Conn, _ *nats.Subscription, err error) {
		slog.Error("error", "error", err)
	})
	nc.SetReconnectHandler(func(_ *nats.Conn) {
		slog.Info("reconnected")
	})
	nc.SetDisconnectErrHandler(func(_ *nats.Conn, err error) {
		slog.Error("disconnected", "error", err)
	})

	return nc, err
}
