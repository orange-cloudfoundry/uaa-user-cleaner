package main

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	gMetricNbErrors = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "uaa_cleaner",
			Name:      "errors",
			Help:      "Number of errors encountered on last user scan",
		},
	)
	gMetricNbUsers = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "uaa_cleaner",
			Name:      "users",
			Help:      "Number users reported on last user scan",
		},
		[]string{"origin"},
	)
	gMetricDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: "uaa_cleaner",
			Name:      "duration",
			Help:      "Duration of the user scan",
			Buckets:   []float64{15, 30, 60, 120, 240, 480},
		},
	)
)

func init() {
	prometheus.MustRegister(gMetricNbErrors)
	prometheus.MustRegister(gMetricNbUsers)
	prometheus.MustRegister(gMetricDuration)
}

// Local Variables:
// ispell-local-dictionary: "american"
// End:
