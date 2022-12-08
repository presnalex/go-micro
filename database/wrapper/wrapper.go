package wrapper

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/prometheus/client_golang/prometheus"
	"go.unistack.org/micro/v3/logger"
)

var (
	DefaultStatsInterval = 5 * time.Second
	// default metric prefix
	DefaultMetricPrefix = "micro_sql_"
	// default label prefix
	DefaultLabelPrefix = "micro_"

	opsCounter           *prometheus.CounterVec
	timeCounterSummary   *prometheus.SummaryVec
	timeCounterHistogram *prometheus.HistogramVec

	mu sync.Mutex

	dbMaxOpenConn       *prometheus.GaugeVec
	dbOpenConn          *prometheus.GaugeVec
	dbInUseConn         *prometheus.GaugeVec
	dbIdleConn          *prometheus.GaugeVec
	dbWaitedConn        *prometheus.GaugeVec
	dbBlockedSeconds    *prometheus.GaugeVec
	dbClosedMaxIdle     *prometheus.GaugeVec
	dbClosedMaxLifetime *prometheus.GaugeVec
)

func registerMetrics() {
	mu.Lock()
	defer mu.Unlock()

	if opsCounter == nil {
		opsCounter = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: fmt.Sprintf("%srequest_total", DefaultMetricPrefix),
				Help: "How many requests processed, partitioned by query and status",
			},
			[]string{
				fmt.Sprintf("%s%s", DefaultLabelPrefix, "dbhost"),
				fmt.Sprintf("%s%s", DefaultLabelPrefix, "dbname"),
				fmt.Sprintf("%s%s", DefaultLabelPrefix, "name"),
				fmt.Sprintf("%s%s", DefaultLabelPrefix, "version"),
				fmt.Sprintf("%s%s", DefaultLabelPrefix, "id"),
				fmt.Sprintf("%s%s", DefaultLabelPrefix, "query"),
				fmt.Sprintf("%s%s", DefaultLabelPrefix, "status"),
			},
		)
	}

	if timeCounterSummary == nil {
		timeCounterSummary = prometheus.NewSummaryVec(
			prometheus.SummaryOpts{
				Name: fmt.Sprintf("%slatency_microseconds", DefaultMetricPrefix),
				Help: "DB request latencies in microseconds, partitioned by query",
			},
			[]string{
				fmt.Sprintf("%s%s", DefaultLabelPrefix, "dbhost"),
				fmt.Sprintf("%s%s", DefaultLabelPrefix, "dbname"),
				fmt.Sprintf("%s%s", DefaultLabelPrefix, "name"),
				fmt.Sprintf("%s%s", DefaultLabelPrefix, "version"),
				fmt.Sprintf("%s%s", DefaultLabelPrefix, "id"),
				fmt.Sprintf("%s%s", DefaultLabelPrefix, "query"),
			},
		)
	}

	if timeCounterHistogram == nil {
		timeCounterHistogram = prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name: fmt.Sprintf("%srequest_duration_seconds", DefaultMetricPrefix),
				Help: "DB request time in seconds, partitioned by query",
			},
			[]string{
				fmt.Sprintf("%s%s", DefaultLabelPrefix, "dbhost"),
				fmt.Sprintf("%s%s", DefaultLabelPrefix, "dbname"),
				fmt.Sprintf("%s%s", DefaultLabelPrefix, "name"),
				fmt.Sprintf("%s%s", DefaultLabelPrefix, "version"),
				fmt.Sprintf("%s%s", DefaultLabelPrefix, "id"),
				fmt.Sprintf("%s%s", DefaultLabelPrefix, "query"),
			},
		)
	}

	if dbMaxOpenConn == nil {
		dbMaxOpenConn = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: fmt.Sprintf("%smax_open_conn", DefaultMetricPrefix),
				Help: "Maximum number of open connections to the database",
			},
			[]string{
				fmt.Sprintf("%s%s", DefaultLabelPrefix, "dbhost"),
				fmt.Sprintf("%s%s", DefaultLabelPrefix, "dbname"),
				fmt.Sprintf("%s%s", DefaultLabelPrefix, "name"),
				fmt.Sprintf("%s%s", DefaultLabelPrefix, "version"),
				fmt.Sprintf("%s%s", DefaultLabelPrefix, "id"),
			},
		)
	}
	if dbOpenConn == nil {
		dbOpenConn = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: fmt.Sprintf("%sopen_conn", DefaultMetricPrefix),
				Help: "The number of established connections both in use and idle",
			},
			[]string{
				fmt.Sprintf("%s%s", DefaultLabelPrefix, "dbhost"),
				fmt.Sprintf("%s%s", DefaultLabelPrefix, "dbname"),
				fmt.Sprintf("%s%s", DefaultLabelPrefix, "name"),
				fmt.Sprintf("%s%s", DefaultLabelPrefix, "version"),
				fmt.Sprintf("%s%s", DefaultLabelPrefix, "id"),
			},
		)
	}
	if dbInUseConn == nil {
		dbInUseConn = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: fmt.Sprintf("%sinuse_conn", DefaultMetricPrefix),
				Help: "The number of connections currently in use",
			},
			[]string{
				fmt.Sprintf("%s%s", DefaultLabelPrefix, "dbhost"),
				fmt.Sprintf("%s%s", DefaultLabelPrefix, "dbname"),
				fmt.Sprintf("%s%s", DefaultLabelPrefix, "name"),
				fmt.Sprintf("%s%s", DefaultLabelPrefix, "version"),
				fmt.Sprintf("%s%s", DefaultLabelPrefix, "id"),
			},
		)
	}
	if dbIdleConn == nil {
		dbIdleConn = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: fmt.Sprintf("%sidle_open", DefaultMetricPrefix),
				Help: "The number of idle connections",
			},
			[]string{
				fmt.Sprintf("%s%s", DefaultLabelPrefix, "dbhost"),
				fmt.Sprintf("%s%s", DefaultLabelPrefix, "dbname"),
				fmt.Sprintf("%s%s", DefaultLabelPrefix, "name"),
				fmt.Sprintf("%s%s", DefaultLabelPrefix, "version"),
				fmt.Sprintf("%s%s", DefaultLabelPrefix, "id"),
			},
		)
	}
	if dbWaitedConn == nil {
		dbWaitedConn = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: fmt.Sprintf("%swaited_conn", DefaultMetricPrefix),
				Help: "The total number of connections waited for",
			},
			[]string{
				fmt.Sprintf("%s%s", DefaultLabelPrefix, "dbhost"),
				fmt.Sprintf("%s%s", DefaultLabelPrefix, "dbname"),
				fmt.Sprintf("%s%s", DefaultLabelPrefix, "name"),
				fmt.Sprintf("%s%s", DefaultLabelPrefix, "version"),
				fmt.Sprintf("%s%s", DefaultLabelPrefix, "id"),
			},
		)
	}
	if dbBlockedSeconds == nil {
		dbBlockedSeconds = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: fmt.Sprintf("%sblocked_seconds", DefaultMetricPrefix),
				Help: "The total time blocked waiting for a new connection",
			},
			[]string{
				fmt.Sprintf("%s%s", DefaultLabelPrefix, "dbhost"),
				fmt.Sprintf("%s%s", DefaultLabelPrefix, "dbname"),
				fmt.Sprintf("%s%s", DefaultLabelPrefix, "name"),
				fmt.Sprintf("%s%s", DefaultLabelPrefix, "version"),
				fmt.Sprintf("%s%s", DefaultLabelPrefix, "id"),
			},
		)
	}
	if dbClosedMaxIdle == nil {
		dbClosedMaxIdle = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: fmt.Sprintf("%sclosed_max_idle", DefaultMetricPrefix),
				Help: "The total number of connections closed due to SetMaxIdleConns",
			},
			[]string{
				fmt.Sprintf("%s%s", DefaultLabelPrefix, "dbhost"),
				fmt.Sprintf("%s%s", DefaultLabelPrefix, "dbname"),
				fmt.Sprintf("%s%s", DefaultLabelPrefix, "name"),
				fmt.Sprintf("%s%s", DefaultLabelPrefix, "version"),
				fmt.Sprintf("%s%s", DefaultLabelPrefix, "id"),
			},
		)
	}
	if dbClosedMaxLifetime == nil {
		dbClosedMaxLifetime = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: fmt.Sprintf("%sclosed_max_lifetime", DefaultMetricPrefix),
				Help: "The total number of connections closed due to SetConnMaxLifetime",
			},
			[]string{
				fmt.Sprintf("%s%s", DefaultLabelPrefix, "dbhost"),
				fmt.Sprintf("%s%s", DefaultLabelPrefix, "dbname"),
				fmt.Sprintf("%s%s", DefaultLabelPrefix, "name"),
				fmt.Sprintf("%s%s", DefaultLabelPrefix, "version"),
				fmt.Sprintf("%s%s", DefaultLabelPrefix, "id"),
			},
		)
	}

	for _, collector := range []prometheus.Collector{
		opsCounter,
		timeCounterSummary,
		timeCounterHistogram,
		dbMaxOpenConn,
		dbOpenConn,
		dbInUseConn,
		dbIdleConn,
		dbWaitedConn,
		dbBlockedSeconds,
		dbClosedMaxIdle,
		dbClosedMaxLifetime,
	} {
		if err := prometheus.DefaultRegisterer.Register(collector); err != nil {
			// if already registered, skip fatal
			if _, ok := err.(prometheus.AlreadyRegisteredError); !ok {
				logger.Fatal(context.Background(), err.Error())
			}
		}
	}
}

type Options struct {
	DBHost  string
	DBName  string
	Name    string
	Version string
	ID      string
}

type Option func(*Options)

func DBHost(dbhost string) Option {
	return func(opts *Options) {
		opts.DBHost = dbhost
	}
}

func DBName(dbname string) Option {
	return func(opts *Options) {
		opts.DBName = dbname
	}
}

func ServiceName(name string) Option {
	return func(opts *Options) {
		opts.Name = name
	}
}

func ServiceVersion(version string) Option {
	return func(opts *Options) {
		opts.Version = version
	}
}

func ServiceID(id string) Option {
	return func(opts *Options) {
		opts.ID = id
	}
}

type Wrapper struct {
	db      *sqlx.DB
	options Options
	labels  []string
}

type TxWrapper struct {
	db      *sqlx.Tx
	options Options
	labels  []string
}

type queryKey struct{}

func QueryContext(ctx context.Context, name string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	return context.WithValue(ctx, queryKey{}, name)
}

func getName(ctx context.Context) string {
	name := "Unknown"

	val, ok := ctx.Value(queryKey{}).(string)
	if ok && len(val) > 0 {
		name = val
	}

	return name
}

func NewWrapper(db *sqlx.DB, opts ...Option) *Wrapper {
	registerMetrics()

	options := Options{}
	for _, opt := range opts {
		opt(&options)
	}

	w := &Wrapper{
		db:      db,
		options: options,
		labels:  []string{options.DBHost, options.DBName, options.Name, options.Version, options.ID},
	}

	go w.collect()

	return w
}

func newTxWrapper(db *sqlx.Tx, opts ...Option) *TxWrapper {
	options := Options{}
	for _, opt := range opts {
		opt(&options)
	}

	w := &TxWrapper{
		db:      db,
		options: options,
	}

	return w
}

func (w *Wrapper) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	var err error

	name := getName(ctx)

	timer := prometheus.NewTimer(prometheus.ObserverFunc(func(v float64) {
		us := v * 1000000 // make microseconds
		timeCounterSummary.WithLabelValues(append(w.labels, name)...).Observe(us)
		timeCounterHistogram.WithLabelValues(append(w.labels, name)...).Observe(v)
	}))
	defer timer.ObserveDuration()

	res, err := w.db.ExecContext(ctx, query, args...)
	if err != nil && err != sql.ErrNoRows {
		opsCounter.WithLabelValues(append(w.labels, name, "failure")...).Inc()
	} else {
		opsCounter.WithLabelValues(append(w.labels, name, "success")...).Inc()
	}

	return res, err
}

func (w *Wrapper) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	var err error

	name := getName(ctx)

	timer := prometheus.NewTimer(prometheus.ObserverFunc(func(v float64) {
		us := v * 1000000 // make microseconds
		timeCounterSummary.WithLabelValues(append(w.labels, name)...).Observe(us)
		timeCounterHistogram.WithLabelValues(append(w.labels, name)...).Observe(v)
	}))
	defer timer.ObserveDuration()

	res, err := w.db.QueryContext(ctx, query, args...)
	if err != nil && err != sql.ErrNoRows {
		opsCounter.WithLabelValues(append(w.labels, name, "failure")...).Inc()
	} else {
		opsCounter.WithLabelValues(append(w.labels, name, "success")...).Inc()
	}

	return res, err
}

func (w *Wrapper) QueryxContext(ctx context.Context, query string, args ...interface{}) (*sqlx.Rows, error) {
	var err error

	name := getName(ctx)

	timer := prometheus.NewTimer(prometheus.ObserverFunc(func(v float64) {
		us := v * 1000000 // make microseconds
		timeCounterSummary.WithLabelValues(append(w.labels, name)...).Observe(us)
		timeCounterHistogram.WithLabelValues(append(w.labels, name)...).Observe(v)
	}))
	defer timer.ObserveDuration()

	res, err := w.db.QueryxContext(ctx, query, args...)
	if err != nil && err != sql.ErrNoRows {
		opsCounter.WithLabelValues(append(w.labels, name, "failure")...).Inc()
	} else {
		opsCounter.WithLabelValues(append(w.labels, name, "success")...).Inc()
	}

	return res, err
}

func (w *Wrapper) QueryRowxContext(ctx context.Context, query string, args ...interface{}) *sqlx.Row {

	name := getName(ctx)

	timer := prometheus.NewTimer(prometheus.ObserverFunc(func(v float64) {
		us := v * 1000000 // make microseconds
		timeCounterSummary.WithLabelValues(append(w.labels, name)...).Observe(us)
		timeCounterHistogram.WithLabelValues(append(w.labels, name)...).Observe(v)
	}))
	defer timer.ObserveDuration()

	res := w.db.QueryRowxContext(ctx, query, args...)
	if res.Err() != nil && res.Err() != sql.ErrNoRows {
		opsCounter.WithLabelValues(append(w.labels, name, "failure")...).Inc()
	} else {
		opsCounter.WithLabelValues(append(w.labels, name, "success")...).Inc()
	}

	return res
}

func (w *Wrapper) GetContext(ctx context.Context, dst interface{}, query string, args ...interface{}) error {
	var err error

	name := getName(ctx)

	timer := prometheus.NewTimer(prometheus.ObserverFunc(func(v float64) {
		us := v * 1000000 // make microseconds
		timeCounterSummary.WithLabelValues(append(w.labels, name)...).Observe(us)
		timeCounterHistogram.WithLabelValues(append(w.labels, name)...).Observe(v)
	}))
	defer timer.ObserveDuration()

	err = w.db.GetContext(ctx, dst, query, args...)
	if err != nil && err != sql.ErrNoRows {
		opsCounter.WithLabelValues(append(w.labels, name, "failure")...).Inc()
	} else {
		opsCounter.WithLabelValues(append(w.labels, name, "success")...).Inc()
	}

	return err
}

func (w *Wrapper) SelectContext(ctx context.Context, dst interface{}, query string, args ...interface{}) error {
	var err error

	name := getName(ctx)

	timer := prometheus.NewTimer(prometheus.ObserverFunc(func(v float64) {
		us := v * 1000000 // make microseconds
		timeCounterSummary.WithLabelValues(append(w.labels, name)...).Observe(us)
		timeCounterHistogram.WithLabelValues(append(w.labels, name)...).Observe(v)
	}))
	defer timer.ObserveDuration()

	err = w.db.SelectContext(ctx, dst, query, args...)
	if err != nil && err != sql.ErrNoRows {
		opsCounter.WithLabelValues(append(w.labels, name, "failure")...).Inc()
	} else {
		opsCounter.WithLabelValues(append(w.labels, name, "success")...).Inc()
	}

	return err
}

func (w *Wrapper) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	var err error

	name := getName(ctx)

	timer := prometheus.NewTimer(prometheus.ObserverFunc(func(v float64) {
		us := v * 1000000 // make microseconds
		timeCounterSummary.WithLabelValues(append(w.labels, name)...).Observe(us)
		timeCounterHistogram.WithLabelValues(append(w.labels, name)...).Observe(v)
	}))
	defer timer.ObserveDuration()

	res, err := w.db.PrepareContext(ctx, query)
	if err != nil {
		opsCounter.WithLabelValues(append(w.labels, name, "failure")...).Inc()
	} else {
		opsCounter.WithLabelValues(append(w.labels, name, "success")...).Inc()
	}

	return res, err
}

func (w *Wrapper) Close() error {
	return w.db.Close()
}

func (w *Wrapper) BeginTxx(ctx context.Context, opts *sql.TxOptions) (*TxWrapper, error) {
	// TODO: now we don't log transaction start/rollback/commit
	res, err := w.db.BeginTxx(ctx, opts)
	return &TxWrapper{db: res, options: w.options, labels: w.labels}, err
}

func (w *TxWrapper) GetContext(ctx context.Context, dst interface{}, query string, args ...interface{}) error {
	var err error

	name := getName(ctx)

	timer := prometheus.NewTimer(prometheus.ObserverFunc(func(v float64) {
		us := v * 1000000 // make microseconds
		timeCounterSummary.WithLabelValues(append(w.labels, name)...).Observe(us)
		timeCounterHistogram.WithLabelValues(append(w.labels, name)...).Observe(v)
	}))
	defer timer.ObserveDuration()

	err = w.db.GetContext(ctx, dst, query, args...)
	if err != nil && err != sql.ErrNoRows {
		opsCounter.WithLabelValues(append(w.labels, name, "failure")...).Inc()
	} else {
		opsCounter.WithLabelValues(append(w.labels, name, "success")...).Inc()
	}

	return err
}

func (w *TxWrapper) QueryRowxContext(ctx context.Context, query string, args ...interface{}) *sqlx.Row {
	var err error

	name := getName(ctx)

	timer := prometheus.NewTimer(prometheus.ObserverFunc(func(v float64) {
		us := v * 1000000 // make microseconds
		timeCounterSummary.WithLabelValues(append(w.labels, name)...).Observe(us)
		timeCounterHistogram.WithLabelValues(append(w.labels, name)...).Observe(v)
	}))
	defer timer.ObserveDuration()

	res := w.db.QueryRowxContext(ctx, query, args...)
	if res.Err() != nil && err != sql.ErrNoRows {
		opsCounter.WithLabelValues(append(w.labels, name, "failure")...).Inc()
	} else {
		opsCounter.WithLabelValues(append(w.labels, name, "success")...).Inc()
	}
	return res
}

func (w *TxWrapper) SelectContext(ctx context.Context, dst interface{}, query string, args ...interface{}) error {
	var err error

	name := getName(ctx)

	timer := prometheus.NewTimer(prometheus.ObserverFunc(func(v float64) {
		us := v * 1000000 // make microseconds
		timeCounterSummary.WithLabelValues(append(w.labels, name)...).Observe(us)
		timeCounterHistogram.WithLabelValues(append(w.labels, name)...).Observe(v)
	}))
	defer timer.ObserveDuration()

	err = w.db.SelectContext(ctx, dst, query, args...)
	if err != nil && err != sql.ErrNoRows {
		opsCounter.WithLabelValues(append(w.labels, name, "failure")...).Inc()
	} else {
		opsCounter.WithLabelValues(append(w.labels, name, "success")...).Inc()
	}

	return err
}

func (w *TxWrapper) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	var err error

	name := getName(ctx)

	timer := prometheus.NewTimer(prometheus.ObserverFunc(func(v float64) {
		us := v * 1000000 // make microseconds
		timeCounterSummary.WithLabelValues(append(w.labels, name)...).Observe(us)
		timeCounterHistogram.WithLabelValues(append(w.labels, name)...).Observe(v)
	}))
	defer timer.ObserveDuration()

	res, err := w.db.ExecContext(ctx, query, args...)
	if err != nil && err != sql.ErrNoRows {
		opsCounter.WithLabelValues(append(w.labels, name, "failure")...).Inc()
	} else {
		opsCounter.WithLabelValues(append(w.labels, name, "success")...).Inc()
	}

	return res, err
}

func (w *TxWrapper) Commit() error {
	return w.db.Commit()
}

func (w *TxWrapper) Rollback() error {
	return w.db.Rollback()
}

func (w *Wrapper) collect() {
	labels := []string{w.options.DBHost, w.options.DBName, w.options.Name, w.options.Version, w.options.ID}

	ticker := time.NewTicker(DefaultStatsInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if w.db == nil {
				continue
			}

			stats := w.db.Stats()
			dbMaxOpenConn.WithLabelValues(labels...).Set(float64(stats.MaxOpenConnections))
			dbOpenConn.WithLabelValues(labels...).Set(float64(stats.OpenConnections))
			dbInUseConn.WithLabelValues(labels...).Set(float64(stats.InUse))
			dbIdleConn.WithLabelValues(labels...).Set(float64(stats.Idle))
			dbWaitedConn.WithLabelValues(labels...).Set(float64(stats.WaitCount))
			dbBlockedSeconds.WithLabelValues(labels...).Set(stats.WaitDuration.Seconds())
			dbClosedMaxIdle.WithLabelValues(labels...).Set(float64(stats.MaxIdleClosed))
			dbClosedMaxLifetime.WithLabelValues(labels...).Set(float64(stats.MaxLifetimeClosed))
		}
	}
}
