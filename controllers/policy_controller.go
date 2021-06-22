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

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

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

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Policy object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *PolicyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	promLabels := prometheus.Labels{
		"name":      req.Name,
		"namespace": req.Namespace,
	}

	pol := &policiesv1.Policy{}
	err := r.Client.Get(ctx, req.NamespacedName, pol)
	if err != nil {
		if errors.IsNotFound(err) {
			metricDeleted := policyStatusMeter.Delete(promLabels)
			logger.Info("Policy not found - must have been deleted.", "metricDeleted", metricDeleted)
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get Policy")
		return ctrl.Result{}, err
	}

	logger.Info("Got ComplianceState", "pol.Status.ComplianceState", pol.Status.ComplianceState)
	if pol.Status.ComplianceState == policiesv1.Compliant {
		metric, err := policyStatusMeter.GetMetricWith(promLabels)
		if err != nil {
			logger.Error(err, "Failed to get metric from GaugeVec")
			return ctrl.Result{}, err
		}
		metric.Set(0)
	} else if pol.Status.ComplianceState == policiesv1.NonCompliant {
		metric, err := policyStatusMeter.GetMetricWith(promLabels)
		if err != nil {
			logger.Error(err, "Failed to get metric from GaugeVec")
			return ctrl.Result{}, err
		}
		metric.Set(1)
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PolicyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&policiesv1.Policy{}).
		Complete(r)
}
