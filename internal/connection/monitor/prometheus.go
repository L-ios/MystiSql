package monitor

import (
	"strconv"

	"MystiSql/internal/connection"

	"github.com/prometheus/client_golang/prometheus"
)

type PrometheusExporter struct {
	collector *Collector

	poolTotalConnections  *prometheus.GaugeVec
	poolIdleConnections   *prometheus.GaugeVec
	poolActiveConnections *prometheus.GaugeVec
	poolMaxConnections    *prometheus.GaugeVec
	poolMinConnections    *prometheus.GaugeVec

	acquireTotal    *prometheus.CounterVec
	acquireFailed   *prometheus.CounterVec
	acquireDuration *prometheus.SummaryVec
	waitDuration    *prometheus.SummaryVec
	waitCount       *prometheus.CounterVec

	queryTotal    *prometheus.CounterVec
	queryFailed   *prometheus.CounterVec
	queryDuration *prometheus.SummaryVec

	execTotal    *prometheus.CounterVec
	execFailed   *prometheus.CounterVec
	execDuration *prometheus.SummaryVec

	healthCheckTotal  *prometheus.CounterVec
	healthCheckFailed *prometheus.CounterVec

	connectionsCreated *prometheus.CounterVec
	connectionsClosed  *prometheus.CounterVec
}

func NewPrometheusExporter(collector *Collector) *PrometheusExporter {
	namespace := "mystisql"
	subsystem := "connection_pool"

	e := &PrometheusExporter{
		collector: collector,

		poolTotalConnections: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "total_connections",
				Help:      "Total number of connections in the pool",
			},
			[]string{"instance"},
		),
		poolIdleConnections: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "idle_connections",
				Help:      "Number of idle connections in the pool",
			},
			[]string{"instance"},
		),
		poolActiveConnections: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "active_connections",
				Help:      "Number of active connections in the pool",
			},
			[]string{"instance"},
		),
		poolMaxConnections: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "max_connections",
				Help:      "Maximum number of connections allowed in the pool",
			},
			[]string{"instance"},
		),
		poolMinConnections: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "min_connections",
				Help:      "Minimum number of connections in the pool",
			},
			[]string{"instance"},
		),

		acquireTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "acquire_total",
				Help:      "Total number of connection acquires",
			},
			[]string{"instance"},
		),
		acquireFailed: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "acquire_failed_total",
				Help:      "Total number of failed connection acquires",
			},
			[]string{"instance"},
		),
		acquireDuration: prometheus.NewSummaryVec(
			prometheus.SummaryOpts{
				Namespace:  namespace,
				Subsystem:  subsystem,
				Name:       "acquire_duration_seconds",
				Help:       "Duration of connection acquire operations",
				Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
			},
			[]string{"instance"},
		),
		waitDuration: prometheus.NewSummaryVec(
			prometheus.SummaryOpts{
				Namespace:  namespace,
				Subsystem:  subsystem,
				Name:       "wait_duration_seconds",
				Help:       "Duration of wait for available connections",
				Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
			},
			[]string{"instance"},
		),
		waitCount: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "wait_total",
				Help:      "Total number of waits for available connections",
			},
			[]string{"instance"},
		),

		queryTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "query_total",
				Help:      "Total number of queries executed",
			},
			[]string{"instance"},
		),
		queryFailed: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "query_failed_total",
				Help:      "Total number of failed queries",
			},
			[]string{"instance"},
		),
		queryDuration: prometheus.NewSummaryVec(
			prometheus.SummaryOpts{
				Namespace:  namespace,
				Subsystem:  subsystem,
				Name:       "query_duration_seconds",
				Help:       "Duration of query execution",
				Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
			},
			[]string{"instance"},
		),

		execTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "exec_total",
				Help:      "Total number of exec operations",
			},
			[]string{"instance"},
		),
		execFailed: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "exec_failed_total",
				Help:      "Total number of failed exec operations",
			},
			[]string{"instance"},
		),
		execDuration: prometheus.NewSummaryVec(
			prometheus.SummaryOpts{
				Namespace:  namespace,
				Subsystem:  subsystem,
				Name:       "exec_duration_seconds",
				Help:       "Duration of exec operations",
				Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
			},
			[]string{"instance"},
		),

		healthCheckTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "health_check_total",
				Help:      "Total number of health checks",
			},
			[]string{"instance"},
		),
		healthCheckFailed: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "health_check_failed_total",
				Help:      "Total number of failed health checks",
			},
			[]string{"instance"},
		),

		connectionsCreated: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "connections_created_total",
				Help:      "Total number of connections created",
			},
			[]string{"instance"},
		),
		connectionsClosed: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "connections_closed_total",
				Help:      "Total number of connections closed",
			},
			[]string{"instance"},
		),
	}

	collector.RegisterEventHandler(e.handleEvent)

	return e
}

func (e *PrometheusExporter) handleEvent(event connection.MetricsEvent) {
	instance := event.Instance
	durationSeconds := float64(event.Duration) / 1e9

	switch event.Type {
	case "acquire":
		e.acquireTotal.WithLabelValues(instance).Inc()
		e.acquireDuration.WithLabelValues(instance).Observe(durationSeconds)
		if !event.Success {
			e.acquireFailed.WithLabelValues(instance).Inc()
		}
	case "release":
	case "wait":
		e.waitCount.WithLabelValues(instance).Inc()
		e.waitDuration.WithLabelValues(instance).Observe(durationSeconds)
	case "query":
		e.queryTotal.WithLabelValues(instance).Inc()
		e.queryDuration.WithLabelValues(instance).Observe(durationSeconds)
		if !event.Success {
			e.queryFailed.WithLabelValues(instance).Inc()
		}
	case "exec":
		e.execTotal.WithLabelValues(instance).Inc()
		e.execDuration.WithLabelValues(instance).Observe(durationSeconds)
		if !event.Success {
			e.execFailed.WithLabelValues(instance).Inc()
		}
	case "health_check":
		e.healthCheckTotal.WithLabelValues(instance).Inc()
		if !event.Success {
			e.healthCheckFailed.WithLabelValues(instance).Inc()
		}
	case "connection_created":
		e.connectionsCreated.WithLabelValues(instance).Inc()
	case "connection_closed":
		e.connectionsClosed.WithLabelValues(instance).Inc()
	}
}

func (e *PrometheusExporter) UpdateMetrics() {
	metrics := e.collector.GetAllMetrics()
	for instance, stats := range metrics {
		e.poolTotalConnections.WithLabelValues(instance).Set(float64(stats.TotalConnections))
		e.poolIdleConnections.WithLabelValues(instance).Set(float64(stats.IdleConnections))
		e.poolActiveConnections.WithLabelValues(instance).Set(float64(stats.ActiveConnections))
		e.poolMaxConnections.WithLabelValues(instance).Set(float64(stats.MaxConnections))
		e.poolMinConnections.WithLabelValues(instance).Set(float64(stats.MinConnections))
	}
}

func (e *PrometheusExporter) MustRegister() {
	prometheus.MustRegister(
		e.poolTotalConnections,
		e.poolIdleConnections,
		e.poolActiveConnections,
		e.poolMaxConnections,
		e.poolMinConnections,
		e.acquireTotal,
		e.acquireFailed,
		e.acquireDuration,
		e.waitDuration,
		e.waitCount,
		e.queryTotal,
		e.queryFailed,
		e.queryDuration,
		e.execTotal,
		e.execFailed,
		e.execDuration,
		e.healthCheckTotal,
		e.healthCheckFailed,
		e.connectionsCreated,
		e.connectionsClosed,
	)
}

func (e *PrometheusExporter) Collect(ch chan<- prometheus.Metric) {
	e.UpdateMetrics()

	e.poolTotalConnections.Collect(ch)
	e.poolIdleConnections.Collect(ch)
	e.poolActiveConnections.Collect(ch)
	e.poolMaxConnections.Collect(ch)
	e.poolMinConnections.Collect(ch)
	e.acquireTotal.Collect(ch)
	e.acquireFailed.Collect(ch)
	e.acquireDuration.Collect(ch)
	e.waitDuration.Collect(ch)
	e.waitCount.Collect(ch)
	e.queryTotal.Collect(ch)
	e.queryFailed.Collect(ch)
	e.queryDuration.Collect(ch)
	e.execTotal.Collect(ch)
	e.execFailed.Collect(ch)
	e.execDuration.Collect(ch)
	e.healthCheckTotal.Collect(ch)
	e.healthCheckFailed.Collect(ch)
	e.connectionsCreated.Collect(ch)
	e.connectionsClosed.Collect(ch)
}

func (e *PrometheusExporter) Describe(ch chan<- *prometheus.Desc) {
	e.poolTotalConnections.Describe(ch)
	e.poolIdleConnections.Describe(ch)
	e.poolActiveConnections.Describe(ch)
	e.poolMaxConnections.Describe(ch)
	e.poolMinConnections.Describe(ch)
	e.acquireTotal.Describe(ch)
	e.acquireFailed.Describe(ch)
	e.acquireDuration.Describe(ch)
	e.waitDuration.Describe(ch)
	e.waitCount.Describe(ch)
	e.queryTotal.Describe(ch)
	e.queryFailed.Describe(ch)
	e.queryDuration.Describe(ch)
	e.execTotal.Describe(ch)
	e.execFailed.Describe(ch)
	e.execDuration.Describe(ch)
	e.healthCheckTotal.Describe(ch)
	e.healthCheckFailed.Describe(ch)
	e.connectionsCreated.Describe(ch)
	e.connectionsClosed.Describe(ch)
}

func FormatStats(instance string, stats *connection.PoolStats) string {
	result := "Connection Pool Stats for " + instance + ":\n"
	result += "  Connections: total=" + strconv.Itoa(stats.TotalConnections) +
		", idle=" + strconv.Itoa(stats.IdleConnections) +
		", active=" + strconv.Itoa(stats.ActiveConnections) +
		", max=" + strconv.Itoa(stats.MaxConnections) +
		", min=" + strconv.Itoa(stats.MinConnections) + "\n"
	result += "  Acquire: total=" + strconv.FormatInt(stats.AcquireCount, 10) +
		", failed=" + strconv.FormatInt(stats.AcquireFailed, 10) +
		", avg_duration=" + strconv.FormatInt(stats.AvgAcquireDuration/1e6, 10) + "ms\n"
	result += "  Wait: count=" + strconv.FormatInt(stats.WaitCount, 10) +
		", max=" + strconv.FormatInt(stats.MaxWaitDuration/1e6, 10) + "ms\n"
	result += "  Query: total=" + strconv.FormatInt(stats.QueryCount, 10) +
		", failed=" + strconv.FormatInt(stats.QueryFailed, 10) + "\n"
	result += "  Exec: total=" + strconv.FormatInt(stats.ExecCount, 10) +
		", failed=" + strconv.FormatInt(stats.ExecFailed, 10) + "\n"
	result += "  Health Checks: total=" + strconv.FormatInt(stats.HealthCheckCount, 10) +
		", failed=" + strconv.FormatInt(stats.HealthCheckFailed, 10) + "\n"

	if stats.LastErrorMsg != "" {
		result += "  Last Error: " + stats.LastErrorMsg + "\n"
	}

	return result
}
