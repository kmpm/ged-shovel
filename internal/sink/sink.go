package sink

import (
	"strings"

	"github.com/kmpm/enp/public/models"
	"github.com/nats-io/nats.go"
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

	return nc.PublishMsg(outbound)
}
