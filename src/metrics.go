package jwtexporter

import "github.com/prometheus/client_golang/prometheus"

const (
	namespace = "jwt_exporter"
)

var (
	// ErrorTotal is a prometheus counter that indicates the total number of unexpected errors encountered by the application
	ErrorTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "error_total",
			Help:      "JWT Exporter Errors",
		},
	)

	JWTExpirySeconds = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "jwt_expires_in_seconds",
			Help:      "Number of seconds until the JWT expires.",
		},
		[]string{"algorithm", "audience", "subject", "id", "scope", "issuer", "secret_key", "secret_name", "secret_namespace", "name", "email", "role"},
	)

	JWTExpirationTimestamp = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "jwt_expiration_timestamp",
			Help:      "Timestamp of when the JWT expires.",
		},
		[]string{"algorithm", "audience", "subject", "id", "scope", "issuer", "secret_key", "secret_name", "secret_namespace", "name", "email", "role"},
	)
	JWTIssuedAtTimestamp = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "jwt_issued_at_timestamp",
			Help:      "Timestamp of when the JWT was issued.",
		},
		[]string{"algorithm", "audience", "subject", "id", "scope", "issuer", "secret_key", "secret_name", "secret_namespace", "name", "email", "role"},
	)
	JWTIssuedSinceSeconds = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "jwt_issued_since_seconds",
			Help:      "Number of seconds since the JWT was issued.",
		},
		[]string{"algorithm", "audience", "subject", "id", "scope", "issuer", "secret_key", "secret_name", "secret_namespace", "name", "email", "role"},
	)
)

func Init() {
	prometheus.MustRegister(ErrorTotal)
	prometheus.MustRegister(JWTExpirySeconds)
	prometheus.MustRegister(JWTExpirationTimestamp)
	prometheus.MustRegister(JWTIssuedAtTimestamp)
	prometheus.MustRegister(JWTIssuedSinceSeconds)
}
