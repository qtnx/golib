package listener

import (
	mainLog "gitlab.id.vin/vincart/golib/log"
	"gitlab.id.vin/vincart/golib/pubsub"
	"gitlab.id.vin/vincart/golib/web/constant"
	"gitlab.id.vin/vincart/golib/web/event"
	"gitlab.id.vin/vincart/golib/web/log"
)

type RequestCompletedLogListener struct {
}

func (r RequestCompletedLogListener) Handler(e pubsub.Event) {
	if e.GetName() != (event.RequestCompletedEvent{}).GetName() {
		return
	}
	payload, ok := e.GetPayload().(event.RequestCompletedPayload)
	if !ok {
		return
	}
	mainLog.Infow([]interface{}{constant.ContextReqMeta, r.makeHttpRequestLog(&payload)}, "finish router")
}

func (r RequestCompletedLogListener) makeHttpRequestLog(message *event.RequestCompletedPayload) *log.HttpRequestLog {
	return &log.HttpRequestLog{
		LoggingContext: log.LoggingContext{
			UserId:            message.UserId,
			DeviceId:          message.DeviceId,
			DeviceSessionId:   message.DeviceSessionId,
			CorrelationId:     message.CorrelationId,
			TechnicalUsername: message.TechnicalUsername,
		},
		Status:         message.Status,
		ExecutionTime:  message.ExecutionTime,
		RequestPattern: message.Mapping,
		RequestPath:    message.Uri,
		Method:         message.Method,
		Query:          message.Query,
		Url:            message.Url,
		RequestId:      message.CorrelationId,
		CallerId:       message.CallerId,
		ClientIp:       message.ClientIpAddress,
		UserAgent:      message.UserAgent,
	}
}
