package events

import (
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
	"strconv"
)

// keys and events in this project
const (
	Payload             = "payload"
	Error               = "error"
	Params              = "params"
	Interconnection     = "interconnection"
	Client              = "client"
	Provider            = "provider"
	MessageIntegrations = "integrationsMessage"
	MessageSalesforce   = "salesforceMessage"
	UserID              = "userID"
	Context             = "context"
	UserContext         = "user.context"
	Message             = "message"

	ContextSaved     = "contextSaved"
	ChatActive       = "chatActive"
	SendMessage      = "sendMessage"
	RetryMessage     = "retryMessage"
	UserBlocked      = "userBlocked"
	StatusSalesforce = "statusSalesforce"
	MessageRepeated  = "messageRepeated"
	MessageSentAgent = "messageSentAgent"
	SendImage        = "sendImage"
)

// GetSpanContextFromSpan returns a SpanContext to be used as parent given a span
func GetSpanContextFromSpan(span ddtrace.Span) ddtrace.SpanContext {
	traceID := strconv.FormatUint(span.Context().TraceID(), 10)
	mapCarrier := tracer.TextMapCarrier{
		tracer.DefaultParentIDHeader: traceID,
		tracer.DefaultTraceIDHeader:  traceID,
	}
	spanContext, _ := tracer.Extract(mapCarrier)
	return spanContext
}
