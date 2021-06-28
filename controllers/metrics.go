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
			"policy_namespace",
		},
	)

	policyDistributedMeter = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ocm_policy_total_managed_clusters",
			Help: "The number of managed clusters the policy is distributed to.",
		},
		[]string{
			"name",
			"policy_namespace",
		},
	)
)

func init() {
	metrics.Registry.MustRegister(
		policyStatusMeter,
		policyDistributedMeter,
	)
}
