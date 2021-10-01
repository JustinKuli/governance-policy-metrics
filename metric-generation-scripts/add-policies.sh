#!/usr/bin/env bash
set -euo pipefail

POLICY_APPLY_NS="${POLICY_APPLY_NS:-default}"
SLEEP_PERIOD="${SLEEP_PERIOD:-75}"
TOTAL_CYCLES="${TOTAL_CYCLES:-50}"
NONCOMPLIANT_FRACTION="${NONCOMPLIANT_FRACTION:-4}"

create_policy () {
    local policy_name="demo-$1"
    local must_have="mustnothave"
    if [[ "$((RANDOM % NONCOMPLIANT_FRACTION))" == "0" ]]; then
        must_have="musthave"
    fi

    echo "Creating $must_have policy named $policy_name"
    kubectl apply -f - <<EOF
apiVersion: policy.open-cluster-management.io/v1
kind: Policy
metadata:
  name: $policy_name
  namespace: $POLICY_APPLY_NS
spec:
  remediationAction: inform
  disabled: false
  policy-templates:
    - objectDefinition:
        apiVersion: policy.open-cluster-management.io/v1
        kind: ConfigurationPolicy
        metadata:
          name: ${policy_name}-ns
        spec:
          remediationAction: inform
          severity: low
          namespaceSelector:
            exclude:
              - kube-*
            include:
              - default
          object-templates:
            - complianceType: ${must_have}
              objectDefinition:
                kind: Namespace
                apiVersion: v1
                metadata:
                  name: ${policy_name}
---
apiVersion: policy.open-cluster-management.io/v1
kind: PlacementBinding
metadata:
  name: binding-$policy_name
  namespace: $POLICY_APPLY_NS
placementRef:
  name: placement-$policy_name
  kind: PlacementRule
  apiGroup: apps.open-cluster-management.io
subjects:
  - name: $policy_name
    kind: Policy
    apiGroup: policy.open-cluster-management.io
---
apiVersion: apps.open-cluster-management.io/v1
kind: PlacementRule
metadata:
  name: placement-$policy_name
  namespace: $POLICY_APPLY_NS
spec:
  clusterConditions:
    - status: 'True'
      type: ManagedClusterConditionAvailable
  clusterSelector:
    matchExpressions: []
EOF
}

CYCLE_COUNT=0
while [[ "$CYCLE_COUNT" -lt "$TOTAL_CYCLES" ]]; do
    CYCLE_COUNT=$((CYCLE_COUNT+1))
    echo "cycle $CYCLE_COUNT / $TOTAL_CYCLES"

    create_policy $CYCLE_COUNT

    sleep "$SLEEP_PERIOD"
done
