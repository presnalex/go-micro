package requestid

import (
	"context"
	"net/textproto"

	"github.com/google/uuid"
	"go.unistack.org/micro/v3/client"
	"go.unistack.org/micro/v3/metadata"
	"go.unistack.org/micro/v3/server"
)

var (
	// default key
	DefaultKey = textproto.CanonicalMIMEHeaderKey("x-request-id")
)

type wrapper struct {
	client.Client
}

func fillOutgoingRequestId(ctx context.Context) (context.Context, error) {
	_, ok := GetOutgoingRequestId(ctx)
	if !ok {
		id, err := uuid.NewRandom()
		if err != nil {
			return ctx, err
		}
		ctx = SetOutgoingRequestId(ctx, id.String())
	}
	return ctx, nil
}

func fillIncomingRequestId(ctx context.Context) (context.Context, error) {
	_, ok := GetIncomingRequestId(ctx)
	if !ok {
		id, err := uuid.NewRandom()
		if err != nil {
			return ctx, err
		}
		ctx = SetIncomingRequestId(ctx, id.String())
	}
	return ctx, nil
}

func generateRequestId() (string, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}

	return id.String(), nil
}

func SetIncomingRequestId(ctx context.Context, requestId string) context.Context {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.New(1)
	}
	md.Set(DefaultKey, requestId)
	metadata.SetIncomingContext(ctx, md)

	return ctx
}

func SetOutgoingRequestId(ctx context.Context, requestId string) context.Context {
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		md = metadata.New(1)
	}
	md.Set(DefaultKey, requestId)
	metadata.SetOutgoingContext(ctx, md)

	return ctx
}

func GetIncomingRequestId(ctx context.Context) (string, bool) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", false
	}
	return md.Get(DefaultKey)
}

func GetOutgoingRequestId(ctx context.Context) (string, bool) {
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		return "", false
	}
	return md.Get(DefaultKey)
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
			var err error
			if ctx, err = fillOutgoingRequestId(ctx); err != nil {
				return err
			}
			return fn(ctx, addr, req, rsp, opts)
		}
	}
}

func (w *wrapper) Call(ctx context.Context, req client.Request, rsp interface{}, opts ...client.CallOption) error {
	var err error
	if ctx, err = fillOutgoingRequestId(ctx); err != nil {
		return err
	}
	return w.Client.Call(ctx, req, rsp, opts...)
}

func (w *wrapper) Stream(ctx context.Context, req client.Request, opts ...client.CallOption) (client.Stream, error) {
	var err error
	if ctx, err = fillOutgoingRequestId(ctx); err != nil {
		return nil, err
	}
	return w.Client.Stream(ctx, req, opts...)
}

func (w *wrapper) Publish(ctx context.Context, p client.Message, opts ...client.PublishOption) error {
	var err error
	if ctx, err = fillOutgoingRequestId(ctx); err != nil {
		return err
	}
	return w.Client.Publish(ctx, p, opts...)
}

func NewServerHandlerWrapper() server.HandlerWrapper {
	return func(fn server.HandlerFunc) server.HandlerFunc {
		return func(ctx context.Context, req server.Request, rsp interface{}) error {
			var err error
			if ctx, err = fillIncomingRequestId(ctx); err != nil {
				return err
			}
			return fn(ctx, req, rsp)
		}
	}
}

func NewServerSubscriberWrapper() server.SubscriberWrapper {
	return func(fn server.SubscriberFunc) server.SubscriberFunc {
		return func(ctx context.Context, msg server.Message) error {
			var err error
			if id, ok := msg.Header()[DefaultKey]; ok {
				ctx = SetIncomingRequestId(ctx, id)
			} else if ctx, err = fillIncomingRequestId(ctx); err != nil {
				return err
			}
			return fn(ctx, msg)
		}
	}
}
