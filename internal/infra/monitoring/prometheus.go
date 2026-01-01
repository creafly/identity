package monitoring

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "identity_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "identity_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	AuthAttemptsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "identity_auth_attempts_total",
			Help: "Total number of authentication attempts",
		},
		[]string{"type", "status"},
	)

	TokensIssuedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "identity_tokens_issued_total",
			Help: "Total number of tokens issued",
		},
		[]string{"type"},
	)

	UsersRegisteredTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "identity_users_registered_total",
			Help: "Total number of registered users",
		},
	)

	ActiveSessions = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "identity_active_sessions",
			Help: "Number of active sessions",
		},
	)

	TenantsTotal = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "identity_tenants_total",
			Help: "Total number of tenants",
		},
	)

	DBQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "identity_db_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
		},
		[]string{"operation"},
	)

	DBConnectionsActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "identity_db_connections_active",
			Help: "Number of active database connections",
		},
	)

	ErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "identity_errors_total",
			Help: "Total number of errors",
		},
		[]string{"type"},
	)
)

func RecordAuthAttempt(authType, status string) {
	AuthAttemptsTotal.WithLabelValues(authType, status).Inc()
}

func RecordTokenIssued(tokenType string) {
	TokensIssuedTotal.WithLabelValues(tokenType).Inc()
}

func RecordUserRegistration() {
	UsersRegisteredTotal.Inc()
}

func RecordError(errorType string) {
	ErrorsTotal.WithLabelValues(errorType).Inc()
}
