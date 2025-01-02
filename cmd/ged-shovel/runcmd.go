package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/kmpm/ged-shovel/internal/eddn"
	"github.com/kmpm/ged-shovel/internal/message"
	"github.com/kmpm/ged-shovel/public/models"

	"github.com/nats-io/jsm.go/natscontext"
	"github.com/nats-io/nats.go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Define Prometheus metrics
var (
	messageDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "ged_shovel_message_duration",
		Help:    "The duration of messages",
		Buckets: []float64{.0001, .00025, .0005, .00075, .001, .0025, .005, .0075, .01},
	}, []string{"status"})

	messagesCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "ged_shovel_messages_per_subject",
		Help: "The number of messages per subject",
	}, []string{"subject", "software", "version"})

	messageSize = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "ged_shovel_message_size",
		Help:    "The size of actual messages excl. headers",
		Buckets: []float64{250, 500, 750, 1000, 1500, 2000, 2500, 5000, 10000, 15000},
	}, []string{"subject"})
)

// Define global variables
var (
	duration time.Duration
	count    float64
)

type RunCmd struct {
	Nats        string `help:"NATS server URI" default:"nats://localhost:4222"`
	NatsContext string `help:"NATS context" default:""`
}

func (cmd *RunCmd) Run() error {
	slog.Info("Running ged-shovel")
	stop := waitfor()
	wg := &sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	validator, err := message.NewValidator()
	if err != nil {
		panic(err)
	}
	slog.Info("validation schemas loaded")

	nc, err := connect(cmd.Nats, cmd.NatsContext)
	if err != nil {
		panic(err)
	}
	slog.Info("connected to nats", "servers", nc.Servers())

	ticker := time.NewTicker(5 * time.Minute)
	go func() {
		for range ticker.C {
			avg := float64(duration.Seconds()) / count
			slog.Info("stats", "count", count, "avg_s", avg)
			count = 0
			duration = 0
		}
	}()
	defer ticker.Stop()

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
				panic(err)
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
		d := time.Since(start)
		messageDuration.WithLabelValues(status).Observe(float64(d.Seconds()))
		if d > 500*time.Millisecond {
			slog.Warn("slow message", "duration", d, "status", status, "schema", schema)
		}
		duration += d
		count++
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
	err = publish(nc, &decoded, rawMsg)
	if err != nil {
		status = "publish_error"
		return fmt.Errorf("error publishing message: %w", err)
	}

	return nil
}

func publish(nc *nats.Conn, inbound *models.EDDN, raw []byte) error {
	// create subject from message schema
	subject := message.Subjectify(inbound.SchemaRef)
	// encode message and compress using zlib
	outbound := nats.NewMsg(subject)
	outbound.Data = raw
	outbound.Header.Add("Content-Encoding", "zlib")
	err := nc.PublishMsg(outbound)
	if err != nil {
		return err
	}
	messageSize.WithLabelValues(subject).Observe(float64(len(inbound.Message)))
	messagesCounter.WithLabelValues(subject, inbound.Header.SoftwareName, inbound.Header.SoftwareVersion).Inc()
	return nil
}

func connect(uri, context string) (nc *nats.Conn, err error) {
	if context != "" {
		nc, err = natscontext.Connect(context, nil)
	} else {
		nc, err = nats.Connect(uri)
	}
	if err != nil {
		return nil, err
	}
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
