package controllers

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	policyStatusGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ocm_policy_status",
			Help: "The compliance status of the named policy. 0 == Compliant. 1 == NonCompliant",
		},
		[]string{
			"type",
			"name",
			"policy_namespace",
			"cluster_namespace",
		},
	)
)

func init() {
	metrics.Registry.MustRegister(
		policyStatusGauge,
	)
}
