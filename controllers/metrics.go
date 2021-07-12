package controllers

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	policyActiveGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ocm_policy_active",
			Help: "Whether the named policy is active. 0 == disabled.",
		},
		[]string{
			"name",
			"policy_namespace",
		},
	)

	policyStatusGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ocm_policy_status",
			Help: "The compliance status of the named policy. 0 == Compliant.",
		},
		[]string{
			"name",
			"policy_namespace",
		},
	)

	// Should match the length of the status.status array in root policies.
	policyDistributedGauge = prometheus.NewGaugeVec(
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
		policyActiveGauge,
		policyStatusGauge,
		policyDistributedGauge,
	)
}
