package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	RequestCount = promauto.NewCounter(prometheus.CounterOpts{
		Name:      "http_requests_total",
		Help:      "The total number of processed requests",
		Namespace: "hund",
	})

	IndexCount = promauto.NewCounter(prometheus.CounterOpts{
		Name:      "generated_indices",
		Help:      "The total number of processed indices",
		Namespace: "hund",
	})

	PackageCount = promauto.NewGauge(prometheus.GaugeOpts{
		Name:      "available_pkgs",
		Help:      "The total number of packages discovered",
		Namespace: "hund",
	})

	ProcessedOutputsCount = promauto.NewCounter(prometheus.CounterOpts{
		Name:      "processed_outputs",
		Help:      "The number of outputs processed",
		Namespace: "hund",
	})

	NixpkgsDate = promauto.NewGauge(prometheus.GaugeOpts{
		Name:      "nixpkgs_date",
		Help:      "The date of the currently used nixpkgs",
		Namespace: "hund",
	})

	LoginAttempts = promauto.NewCounter(prometheus.CounterOpts{
		Name:      "login_attempts",
		Help:      "The total number of login attempts",
		Namespace: "hund",
	})

	RegisterAttempts = promauto.NewCounter(prometheus.CounterOpts{
		Name:      "register_attempts",
		Help:      "The total number of register attempts",
		Namespace: "hund",
	})
)
