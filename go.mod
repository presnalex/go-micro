module github.com/presnalex/go-micro/v3

go 1.16

replace github.com/unistack-org/micro/v3 => go.unistack.org/micro/v3 v3.10.10

require (
	github.com/godror/godror v0.36.0
	github.com/golang/protobuf v1.5.2
	github.com/google/uuid v1.3.0
	github.com/gorilla/mux v1.8.0
	github.com/jackc/pgx/v4 v4.17.2
	github.com/jmoiron/sqlx v1.3.5
	github.com/presnalex/codec-bytes v0.0.1
	github.com/presnalex/micro-wrapper-metrics-prometheus v0.0.1
	github.com/prometheus/client_golang v1.11.1
	github.com/segmentio/encoding v0.3.6
	github.com/twmb/franz-go v1.11.5
	github.com/unistack-org/micro/v3 v3.0.0-gamma.0.20201012090909-6e43ae719076
	go.uber.org/zap v1.19.1
	go.unistack.org/micro-broker-kgo/v3 v3.8.3
	go.unistack.org/micro-codec-proto/v3 v3.10.0
	go.unistack.org/micro-logger-zap/v3 v3.8.0
	go.unistack.org/micro/v3 v3.10.14
	google.golang.org/protobuf v1.28.1
	gopkg.in/yaml.v2 v2.4.0 // indirect
)
