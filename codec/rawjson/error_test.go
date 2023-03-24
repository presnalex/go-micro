package rawjson_test

import (
	"context"
	"testing"

	kgo "go.unistack.org/micro-broker-kgo/v3"
	"go.unistack.org/micro/v3/broker"

	rawjson "github.com/presnalex/go-micro/v3/codec/rawjson"
)

func TestErrorHandler(t *testing.T) {
	ctx := context.Background()
	brk := kgo.NewBroker(broker.Addrs("172.18.0.121:9092"),
		broker.Codec(rawjson.NewCodec()),
	)
	if err := brk.Init(); err != nil {
		t.Fatal(err)
	}
	if err := brk.Connect(ctx); err != nil {
		t.Fatal(err)
	}

	msg := &broker.Message{
		Header: map[string]string{
			"Content-Type":       "application/json",
			"Micro-From-Service": "test",
			"Micro-Topic":        "test_topic",
			"Message-Id":         "1234567890",
		},
		Body: []byte(`{"ACTION": "NEW","ENTITY": false,"EVENTID": "2020061515015270469","ID": "220661547858"}`),
	}
	if err := brk.Publish(ctx, "test_topic", msg); err != nil {
		t.Fatal(err)
	}
}
