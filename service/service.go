package service

import (
	"time"

	"github.com/google/uuid"
	"github.com/presnalex/go-micro/v3/codec/rawjson"
	logwrapper "github.com/presnalex/go-micro/v3/wrapper/logwrapper"
	idwrapper "github.com/presnalex/go-micro/v3/wrapper/requestid"
	promwrapper "github.com/presnalex/micro-wrapper-metrics-prometheus"
	kbroker "go.unistack.org/micro-broker-kgo/v3"
	cp "go.unistack.org/micro-codec-proto/v3"
	"go.unistack.org/micro/v3/broker"
	"go.unistack.org/micro/v3/client"
	"go.unistack.org/micro/v3/server"
)

type MetricConfig struct {
	Addr string `json:"addr" env:"METRIC_ADDRESS"`
}

type ConsulConfig struct {
	Addr          string `json:"addr" env:"consul_host"`
	Token         string `json:"token" env:"consul_config_acl_token"`
	AppPath       string `json:"path" env:"consul_app_path"`
	NamespacePath string `json:"config_path" env:"consul_config_path"`
	WatchWait     string `json:"watch_wait"`
	WatchTimeout  string `json:"watch_timeout"`
}

type VaultConfig struct {
	Uri           string `json:"uri" env:"vault_uri"`
	NamespacePath string `json:"path" env:"vault_config_path"`
	AppPath       string `json:"apppath" env:"vault_app_path"`
	Token         string `json:"token" env:"vault_token"`
	SecretId      string `json:"secretid" env:"vault_secret_id"`
	RoleId        string `json:"roleid" env:"vault_role_id"`
}

type PostgresConfig struct {
	Addr            string `json:"addr"`
	Login           string `json:"login"`
	Passw           string `json:"passw"`
	DBName          string `json:"dbname"`
	AppName         string `json:"appname"`
	ConnMax         int    `json:"conn_max"`
	ConnMaxIdle     int    `json:"conn_max_idle"`
	ConnLifetime    int    `json:"conn_lifetime"`
	ConnMaxIdleTime int    `json:"conn_maxidletime"`
}

type OracleConfig struct {
	Addr            string `json:"addr"`
	Login           string `json:"login"`
	Passw           string `json:"passw"`
	DBName          string `json:"dbname"`
	ConnMax         int    `json:"conn_max"`
	ConnMaxIdle     int    `json:"conn_max_idle"`
	ConnLifetime    int    `json:"conn_lifetime"`
	ConnMaxIdleTime int    `json:"conn_maxidletime"`
}

type ServerConfig struct {
	Name    string `json:"name" env:"SERVER_NAME"`
	ID      string `json:"id" env:"SERVER_ID"`
	Version string `json:"version" env:"SERVER_VERSION"`
	Addr    string `json:"addr" env:"SERVER_ADDRESS"`
}

type CoreConfig struct {
	GetConfig bool `json:"-"`
	Profile   bool `json:"profile"`
}

type ClientConfig struct {
	ClientRetries        int `json:"client_retries"`
	ClientRequestTimeout int `json:"client_request_timeout"`
	ClientPoolSize       int `json:"client_pool_size"`
	ClientDialTimeout    int `json:"client_dial_timeout"`
	ClientPoolTTL        int `json:"client_pool_ttl"`
	TransportTimeout     int `json:"transport_timeout"`
}

func ClientOptions(ccfg *ClientConfig) ([]client.Option, error) {
	clientRetries := ccfg.ClientRetries
	clientPoolSize := ccfg.ClientPoolSize

	clientPoolTTL := time.Duration(ccfg.ClientPoolTTL) * time.Second
	clientDialTimeout := time.Duration(ccfg.ClientDialTimeout) * 5 * time.Second
	clientRequestTimeout := time.Duration(ccfg.ClientRequestTimeout) * 5 * time.Second

	if clientRetries == 0 {
		clientRetries = defaultClientRetries
	}
	if clientPoolSize == 0 {
		clientPoolSize = defaultClientPoolSize
	}
	if ccfg.ClientRequestTimeout == 0 {
		clientRequestTimeout = defaultClientRequestTimeout
	}
	if ccfg.ClientRequestTimeout == 0 {
		clientRequestTimeout = defaultClientRequestTimeout
	}
	if ccfg.ClientPoolTTL == 0 {
		clientPoolTTL = defaultClientPoolTTL
	}
	if ccfg.ClientDialTimeout == 0 {
		clientDialTimeout = defaultClientDialTimeout
	}

	opts := []client.Option{
		client.Codec("application/grpc+proto", cp.NewCodec()),
		client.Codec("application/json", rawjson.NewCodec()),
		// grpc.AuthTLS(&tls.Config{InsecureSkipVerify: true}),
		client.Broker(broker.DefaultBroker),
		client.Retries(clientRetries),
		client.RequestTimeout(clientRequestTimeout),
		client.PoolSize(clientPoolSize),
		client.PoolTTL(clientPoolTTL),
		client.DialTimeout(clientDialTimeout),
		client.Wrap(promwrapper.NewClientWrapper()),
		client.Wrap(idwrapper.NewClientWrapper()),
		client.Wrap(logwrapper.NewClientWrapper()),
	}

	return opts, nil
}

func ServerOptions(scfg *ServerConfig) ([]server.Option, error) {
	if len(scfg.ID) == 0 {
		uid, err := uuid.NewRandom()
		if err != nil {
			return nil, err
		}
		scfg.ID = uid.String()
	}

	opts := []server.Option{
		// sgrpc.AuthTLS(&tls.Config{InsecureSkipVerify: true}),
		server.Name(scfg.Name),
		server.Version(scfg.Version),
		server.Address(scfg.Addr),
		server.ID(scfg.ID),
		//		server.Wait(true),
		server.RegisterTTL(defaultRegisterTTL),
		server.RegisterInterval(defaultRegisterInterval),
		server.Broker(broker.DefaultBroker),
		server.Codec("application/grpc", cp.NewCodec()),
		server.Codec("application/grpc+proto", cp.NewCodec()),
		server.Codec("application/json", rawjson.NewCodec()),
		server.WrapHandler(
			promwrapper.NewHandlerWrapper(
				promwrapper.ServiceName(scfg.Name),
				promwrapper.ServiceVersion(scfg.Version),
				promwrapper.ServiceID(scfg.ID),
			),
		),
		server.WrapHandler(idwrapper.NewServerHandlerWrapper()),
		server.WrapHandler(logwrapper.NewServerHandlerWrapper()),
		server.WrapSubscriber(
			promwrapper.NewSubscriberWrapper(
				promwrapper.ServiceName(scfg.Name),
				promwrapper.ServiceVersion(scfg.Version),
				promwrapper.ServiceID(scfg.ID),
			),
		),
		server.WrapSubscriber(idwrapper.NewServerSubscriberWrapper()),
		server.WrapSubscriber(logwrapper.NewServerSubscriberWrapper()),
	}

	return opts, nil
}

func InitBroker(cfg *BrokerConfig) broker.Broker {
	if cfg == nil {
		return broker.DefaultBroker
	}

	switch cfg.Type {
	case "kafka":

		opts := []broker.Option{broker.Codec(rawjson.NewCodec()), broker.Addrs(cfg.Addr...)}
		opts = append(opts, NewKafkaReaderConfig(cfg)...)
		opts = append(opts, NewKafkaWriterConfig(cfg)...)

		return kbroker.NewBroker(opts...)
	// case "kubemq":
	//	return kubemqbroker.NewBroker(
	//		broker.Addrs(cfg.Addr...),
	//		broker.Codec(rawjson.Marshaler{}),
	//	)
	default:
		return broker.DefaultBroker
	}
}
