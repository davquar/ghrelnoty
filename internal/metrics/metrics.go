package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const namespace = "ghrelnoty"

var dbOpenErrorsCounter = promauto.NewCounter(prometheus.CounterOpts{
	Namespace: namespace,
	Name:      "db_open_errors_total",
	Help:      "Total number of databse open errors",
})

var dbErrorsCounter = promauto.NewCounter(prometheus.CounterOpts{
	Namespace: namespace,
	Name:      "db_errors_total",
	Help:      "Total number of databse errors",
})

var rateLimitRiskCounter = promauto.NewCounter(prometheus.CounterOpts{
	Namespace: namespace,
	Name:      "rate_limit_risks_total",
	Help:      "Total times there was a risk of hitting rate limits",
})

var rateLimitCounter = promauto.NewCounter(prometheus.CounterOpts{
	Namespace: namespace,
	Name:      "rate_limited_total",
	Help:      "Total times rate limits were hit",
})

var rateLimitGauge = promauto.NewGauge(prometheus.GaugeOpts{
	Namespace: namespace,
	Name:      "github_rate_limit",
	Help:      "Value of GitHub's rate limit",
})

var rateLimitUsedGauge = promauto.NewGauge(prometheus.GaugeOpts{
	Namespace: namespace,
	Name:      "github_rate_limit_used",
	Help:      "Current usage of GitHub's rate limit",
})

var releaseGetErrorsCounter = promauto.NewCounter(prometheus.CounterOpts{
	Namespace: namespace,
	Name:      "release_get_errors_total",
	Help:      "Total times it was not possible to get the latest release",
})

var newReleaseFoundCounter = promauto.NewCounter(prometheus.CounterOpts{
	Namespace: namespace,
	Name:      "new_releases_founds_total",
	Help:      "Total times a new release was found",
})

var notificationErrorsCounter = promauto.NewCounter(prometheus.CounterOpts{
	Namespace: namespace,
	Name:      "notification_errors_total",
	Help:      "Total times there were problems notifying",
})

func DBOpenError() {
	dbOpenErrorsCounter.Inc()
}

func DBError() {
	dbErrorsCounter.Inc()
}

func RateLimitRisk() {
	rateLimitRiskCounter.Inc()
}

func RateLimited() {
	rateLimitCounter.Inc()
}

func SetRateLimitValue(value float64) {
	rateLimitGauge.Set(value)
}

func SetRateLimitUsedValue(value float64) {
	rateLimitUsedGauge.Set(value)
}

func CannotGetRelease() {
	releaseGetErrorsCounter.Inc()
}

func NewReleaseFound() {
	newReleaseFoundCounter.Inc()
}

func NotificationError() {
	notificationErrorsCounter.Inc()
}
