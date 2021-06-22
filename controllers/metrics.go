package controllers

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	policyStatusMeter = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ocm_policy_status",
			Help: "The compliance status of the named policy. 0 == Compliant.",
		},
		[]string{
			"name",
			"namespace",
		},
	)
)

func init() {
	metrics.Registry.MustRegister(policyStatusMeter)
}
