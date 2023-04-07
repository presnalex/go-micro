package service

import (
	"context"
	"fmt"
	"time"

	raw "github.com/presnalex/codec-bytes"
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/sasl/plain"
	kafka "go.unistack.org/micro-broker-kgo/v3"
	"go.unistack.org/micro/v3/broker"
	"go.unistack.org/micro/v3/client"
	"go.unistack.org/micro/v3/logger"
	"go.unistack.org/micro/v3/metadata"
)

func NewKafkaWriterConfig(cfg *BrokerConfig) []broker.Option {
	var opts []broker.Option

	kopts := []kgo.Opt{
		kgo.FetchMaxWait(1 * time.Second),
		kgo.StopProducerOnDataLossDetected(),
		kgo.ClientID(cfg.ClientID),
		kgo.ProducerBatchCompression(kgo.NoCompression()),
		kgo.MaxBufferedRecords(1000),
		kgo.RequiredAcks(kgo.LeaderAck()),
		// kgo.ProducerLinger(1 * time.Second), dont set by default to speedup publishing
		kgo.ProducerBatchMaxBytes(1 * 1024 * 1024),
	}
	if len(cfg.Username) > 0 && len(cfg.Password) > 0 {
		kopts = append(kopts,
			kgo.SASL((plain.Auth{User: cfg.Username, Pass: cfg.Password}).AsMechanism()),
		)
	}

	if cfg.Writer.BatchBytes > 0 {
		kopts = append(kopts, kgo.ProducerBatchMaxBytes(int32(cfg.Writer.BatchBytes)))
	}
	if cfg.Writer.BatchTimeout.Duration > 0 {
		kopts = append(kopts, kgo.ProducerLinger(cfg.Writer.BatchTimeout.Duration))
	}

	opts = append(opts,
		kafka.Options(kopts...),
	)

	return opts
}

func NewKafkaReaderConfig(cfg *BrokerConfig) []broker.Option {
	var opts []broker.Option

	kopts := []kgo.Opt{
		kgo.FetchMaxWait(1 * time.Second),
		kgo.ClientID(cfg.ClientID),
		// kgo.AllowedConcurrentFetches(3),
	}
	if len(cfg.Username) > 0 && len(cfg.Password) > 0 {
		kopts = append(kopts,
			kgo.SASL((plain.Auth{User: cfg.Username, Pass: cfg.Password}).AsMechanism()),
		)
	}

	commitInterval := 1 * time.Second

	if cfg.Reader.MinBytes > 0 {
		kopts = append(kopts, kgo.FetchMinBytes(int32(cfg.Reader.MinBytes)))
	}
	if cfg.Reader.MaxBytes > 0 {
		kopts = append(kopts, kgo.FetchMaxBytes(int32(cfg.Reader.MaxBytes)))
		//	kopts = append(kopts, kgo.FetchMaxPartitionBytes(int32(cfg.Reader.MaxBytes)))
		// kopts = append(kopts, kgo.BrokerMaxReadBytes(2*int32(cfg.Reader.MaxBytes)))
	}
	if cfg.Reader.MaxWait.Duration > 0 {
		kopts = append(kopts, kgo.FetchMaxWait(cfg.Reader.MaxWait.Duration))
	}
	if cfg.Reader.CommitInterval.Duration > 0 {
		commitInterval = cfg.Reader.CommitInterval.Duration
	}

	opts = append(opts,
		kafka.CommitInterval(commitInterval),
		kafka.Options(kopts...),
	)

	return opts
}

type BrokerConfig struct {
	Type     string   `json:"type"`
	Workers  int      `json:"workers"`
	Addr     []string `json:"addr"`
	ClientID string   `json:"clientid"`
	Username string   `json:"username"`
	Password string   `json:"password"`
	Reader   struct {
		Group                  string   `json:"group"`
		QueueCapacity          int      `json:"queue_capacity"`
		MinBytes               int      `json:"min_bytes"`
		MaxBytes               int      `json:"max_bytes"`
		MaxWait                Duration `json:"max_wait"`
		ReadLagInterval        Duration `json:"read_lag_interval"`
		HeartbeatInterval      Duration `json:"heartbeat_interval"`
		CommitInterval         Duration `json:"commit_interval"`
		PartitionWatchInterval Duration `json:"partition_watch_interval"`
		SessionTimeout         Duration `json:"session_timeout"`
		RebalanceTimeout       Duration `json:"rebalance_timeout"`
		JoinGroupBackoff       Duration `json:"join_group_backoff"`
		RetentionTime          Duration `json:"retention_time"`
		StartOffset            int64    `json:"offset"`
		ReadBackoffMin         Duration `json:"read_backoff_min"`
		ReadBackoffMax         Duration `json:"read_backoff_max"`
		MaxAttempts            int      `json:"max_attempts"`
	} `json:"reader"`
	Writer struct {
		MaxAttempts  int      `json:"max_attempts"`
		BatchSize    int      `json:"batch_size"`
		BatchBytes   int      `json:"batch_bytes"`
		BatchTimeout Duration `json:"batch_timeout"`
		ReadTimeout  Duration `json:"read_timeout"`
		WriteTimeout Duration `json:"write_timeout"`
		RequiredAcks int      `json:"required_acks"`
	} `json:"writer"`
	Group string `json:"group"`
}

func NewErrorHandler(topic string, appName string, c client.Client) broker.Handler {
	return func(evt broker.Event) error {
		logger.Error(context.Background(), "broken message: %s", evt.Error())

		msg := evt.Message()
		if evt.Error() != nil {
			msg.Header.Set("Micro-Error", fmt.Sprintf("%v", evt.Error()))
		}
		if appName != "" {
			msg.Header.Set("Micro-Appname", appName)
		}

		// create context for wrappers from message header
		ctx := metadata.NewOutgoingContext(context.Background(), msg.Header)

		if err := c.Publish(ctx, c.NewMessage(topic, &raw.Frame{Data: msg.Body}, client.WithMessageContentType("application/json"))); err != nil {
			logger.Fatal(ctx, "cannot publish to %s topic: %s", topic, err)
		}

		if err := evt.Ack(); err != nil {
			logger.Fatal(ctx, "unable to ack broker message: %s", err)
		}

		return evt.Error()
	}
}
