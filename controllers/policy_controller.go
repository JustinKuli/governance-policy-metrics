/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"strings"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	clusterv1 "github.com/open-cluster-management/api/cluster/v1"
	policiesv1 "github.com/open-cluster-management/governance-policy-propagator/pkg/apis/policy/v1"
	"github.com/prometheus/client_golang/prometheus"
)

// PolicyReconciler reconciles a Policy object
type PolicyReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=policy.open-cluster-management.io,resources=policies,verbs=get;list;watch
//+kubebuilder:rbac:groups=policy.open-cluster-management.io,resources=policies/status,verbs=get
//+kubebuilder:rbac:groups=policy.open-cluster-management.io,resources=policies/finalizers,verbs=update
//+kubebuilder:rbac:groups=cluster.open-cluster-management.io,resources=managedclusters,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *PolicyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Need to know if the policy is a root policy to create the correct prometheus labels
	clusterList := &clusterv1.ManagedClusterList{}
	err := r.Client.List(ctx, clusterList, &client.ListOptions{})
	if err != nil {
		logger.Error(err, "Failed to list clusters, going to retry...")
		return ctrl.Result{}, err
	}

	var promLabels map[string]string
	if isInClusterNamespace(req.Namespace, clusterList.Items) {
		// propagated policies look like <namespace>.<name>
		// also note: k8s namespace names follow RFC 1123 (so no "." in it)
		splitName := strings.SplitN(req.Name, ".", 2)
		promLabels = prometheus.Labels{
			"type":              "propagated",
			"name":              splitName[1],
			"policy_namespace":  splitName[0],
			"cluster_namespace": req.Namespace,
		}
	} else {
		promLabels = prometheus.Labels{
			"type":              "root",
			"name":              req.Name,
			"policy_namespace":  req.Namespace,
			"cluster_namespace": "<null>", // this is basically a sentinel value
		}
	}

	pol := &policiesv1.Policy{}
	err = r.Client.Get(ctx, req.NamespacedName, pol)
	if err != nil {
		if errors.IsNotFound(err) {
			// Try to delete the gauge, but don't get hung up on errors.
			statusGaugeDeleted := policyStatusGauge.Delete(promLabels)
			logger.Info("Policy not found - must have been deleted.",
				"status-gauge-deleted", statusGaugeDeleted)
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get Policy")
		return ctrl.Result{}, err
	}

	logger.Info("Got active state", "pol.Spec.Disabled", pol.Spec.Disabled)
	if pol.Spec.Disabled {
		// The policy is no longer active, so delete its metric
		statusGaugeDeleted := policyStatusGauge.Delete(promLabels)
		logger.Info("Metric removed for non-active policy",
			"status-gauge-deleted", statusGaugeDeleted)
		return ctrl.Result{}, nil
	}

	logger.Info("Got ComplianceState", "pol.Status.ComplianceState", pol.Status.ComplianceState)
	statusMetric, err := policyStatusGauge.GetMetricWith(promLabels)
	if err != nil {
		logger.Error(err, "Failed to get status metric from GaugeVec")
		return ctrl.Result{}, err
	}
	if pol.Status.ComplianceState == policiesv1.Compliant {
		statusMetric.Set(0)
	} else if pol.Status.ComplianceState == policiesv1.NonCompliant {
		statusMetric.Set(1)
	}

	return ctrl.Result{}, nil
}

// This would be from a common pkg, but there is a dependency mismatch causing problems.
func isInClusterNamespace(ns string, allClusters []clusterv1.ManagedCluster) bool {
	for _, cluster := range allClusters {
		if ns == cluster.GetName() {
			return true
		}
	}
	return false
}

// SetupWithManager sets up the controller with the Manager.
func (r *PolicyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&policiesv1.Policy{}).
		Complete(r)
}
