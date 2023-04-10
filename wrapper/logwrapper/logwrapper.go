package logwrapper

import (
	"context"

	"github.com/presnalex/go-micro/v3/wrapper/requestid"

	"go.unistack.org/micro/v3/client"
	"go.unistack.org/micro/v3/server"

	"github.com/presnalex/go-micro/v3/logger"
)

type wrapper struct {
	client.Client
}

func NewClientWrapper() client.Wrapper {
	return func(c client.Client) client.Client {
		handler := &wrapper{
			Client: c,
		}
		return handler
	}
}

func NewClientCallWrapper() client.CallWrapper {
	return func(fn client.CallFunc) client.CallFunc {
		return func(ctx context.Context, addr string, req client.Request, rsp interface{}, opts client.CallOptions) error {
			if id, ok := requestid.GetOutgoingRequestId(ctx); ok {
				logger.InjectLogger(ctx, id)
			}
			return fn(ctx, addr, req, rsp, opts)
		}
	}
}

func (w *wrapper) Call(ctx context.Context, req client.Request, rsp interface{}, opts ...client.CallOption) error {
	if id, ok := requestid.GetOutgoingRequestId(ctx); ok {
		logger.InjectLogger(ctx, id)
	}
	return w.Client.Call(ctx, req, rsp, opts...)
}

func (w *wrapper) Stream(ctx context.Context, req client.Request, opts ...client.CallOption) (client.Stream, error) {
	if id, ok := requestid.GetOutgoingRequestId(ctx); ok {
		logger.InjectLogger(ctx, id)
	}
	return w.Client.Stream(ctx, req, opts...)
}

func (w *wrapper) Publish(ctx context.Context, p client.Message, opts ...client.PublishOption) error {
	if id, ok := requestid.GetOutgoingRequestId(ctx); ok {
		logger.InjectLogger(ctx, id)
	}
	return w.Client.Publish(ctx, p, opts...)
}

func NewServerHandlerWrapper() server.HandlerWrapper {
	return func(fn server.HandlerFunc) server.HandlerFunc {
		return func(ctx context.Context, req server.Request, rsp interface{}) error {
			if id, ok := requestid.GetIncomingRequestId(ctx); ok {
				logger.InjectLogger(ctx, id)
			}
			return fn(ctx, req, rsp)
		}
	}
}

func NewServerSubscriberWrapper() server.SubscriberWrapper {
	return func(fn server.SubscriberFunc) server.SubscriberFunc {
		return func(ctx context.Context, msg server.Message) error {
			if id, ok := requestid.GetIncomingRequestId(ctx); ok {
				logger.InjectLogger(ctx, id)
			}
			return fn(ctx, msg)
		}
	}
}
