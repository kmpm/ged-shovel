package sink

import (
	"strings"

	"github.com/kmpm/ged-shovel/public/models"
	"github.com/nats-io/nats.go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	messagesPerSubject = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "ged_shovel_messages_per_subject",
		Help: "The number of messages per subject",
	}, []string{"subject"})
)

func subjectify(schemaRef string) string {
	//schema is a url, so we need to remove schema and the slashes and replace them with dots
	//e.g. https://eddn.edcd.io/schemas/journal/1.json -> eddn.journal.1
	parts := strings.Split(schemaRef, "/")
	return "eddn." + strings.Join(parts[4:], ".")
}

func Publish(nc *nats.Conn, inbound *models.EDDN, raw []byte) error {
	// create subject from message schema
	subject := subjectify(inbound.SchemaRef)
	// encode message and compress using zlib
	outbound := nats.NewMsg(subject)
	outbound.Data = raw
	outbound.Header.Add("Content-Encoding", "zlib")
	messagesPerSubject.WithLabelValues(subject).Inc()
	return nc.PublishMsg(outbound)
}
