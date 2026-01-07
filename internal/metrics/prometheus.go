package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP metrics
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	// Authentication metrics
	LoginAttemptsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "login_attempts_total",
			Help: "Total number of login attempts",
		},
		[]string{"tenant_id", "status"},
	)

	LoginSuccessTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "login_success_total",
			Help: "Total number of successful logins",
		},
	)

	LoginFailureTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "login_failure_total",
			Help: "Total number of failed logins",
		},
	)

	// MFA metrics
	MFAEnrollmentsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "mfa_enrollments_total",
			Help: "Total number of MFA enrollments",
		},
	)

	MFAVerificationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mfa_verifications_total",
			Help: "Total number of MFA verifications",
		},
		[]string{"status"},
	)

	// User management metrics
	UsersCreatedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "users_created_total",
			Help: "Total number of users created",
		},
	)

	UsersActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "users_active",
			Help: "Current number of active users",
		},
	)

	// Database metrics
	DatabaseConnectionsActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "database_connections_active",
			Help: "Current number of active database connections",
		},
	)

	DatabaseQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "database_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
		[]string{"operation"},
	)

	// Cache metrics
	CacheHitsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_hits_total",
			Help: "Total number of cache hits",
		},
		[]string{"cache_type"},
	)

	CacheMissesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_misses_total",
			Help: "Total number of cache misses",
		},
		[]string{"cache_type"},
	)

	// Rate limiting metrics
	RateLimitHitsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rate_limit_hits_total",
			Help: "Total number of rate limit hits",
		},
		[]string{"scope"},
	)
)

