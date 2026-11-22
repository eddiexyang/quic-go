package qlog

import (
	"context"

	"github.com/eddiexyang/quic-go"
	"github.com/eddiexyang/quic-go/qlog"
	"github.com/eddiexyang/quic-go/qlogwriter"
)

const EventSchema = "urn:ietf:params:qlog:events:http3-12"

func DefaultConnectionTracer(ctx context.Context, isClient bool, connID quic.ConnectionID) qlogwriter.Trace {
	return qlog.DefaultConnectionTracerWithSchemas(ctx, isClient, connID, []string{qlog.EventSchema, EventSchema})
}
