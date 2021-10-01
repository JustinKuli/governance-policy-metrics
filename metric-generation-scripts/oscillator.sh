#!/usr/bin/env bash
set -euo pipefail

POLICY_APPLY_NS="${POLICY_APPLY_NS:-default}"
SLEEP_PERIOD="${SLEEP_PERIOD:-90}"
TOTAL_CYCLES="${TOTAL_CYCLES:-20}"
INITIAL_DELAY="${INITIAL_DELAY:-300}"

OSC_ONE="mustnothave"
OSC_TWO="musthave"

echo "Applying mypolicy-1..."
kubectl apply -f - <<EOF
apiVersion: policy.open-cluster-management.io/v1
kind: Policy
metadata:
  name: mypolicy-1
  namespace: $POLICY_APPLY_NS
spec:
  remediationAction: inform
  disabled: false
  policy-templates:
    - objectDefinition:
        apiVersion: policy.open-cluster-management.io/v1
        kind: ConfigurationPolicy
        metadata:
          name: policy-namespace-oscillator-1
        spec:
          remediationAction: inform
          severity: low
          namespaceSelector:
            exclude:
              - kube-*
            include:
              - default
          object-templates:
            - complianceType: $OSC_ONE
              objectDefinition:
                kind: Namespace
                apiVersion: v1
                metadata:
                  name: nonexistentnamespace
---
apiVersion: policy.open-cluster-management.io/v1
kind: PlacementBinding
metadata:
  name: binding-policy-namespace-oscillator-1
  namespace: $POLICY_APPLY_NS
placementRef:
  name: placement-policy-namespace-oscillator
  kind: PlacementRule
  apiGroup: apps.open-cluster-management.io
subjects:
  - name: mypolicy-1
    kind: Policy
    apiGroup: policy.open-cluster-management.io
---
apiVersion: apps.open-cluster-management.io/v1
kind: PlacementRule
metadata:
  name: placement-policy-namespace-oscillator
  namespace: $POLICY_APPLY_NS
spec:
  clusterConditions:
    - status: 'True'
      type: ManagedClusterConditionAvailable
  clusterSelector:
    matchExpressions: []
EOF

echo "Waiting initial delay for mypolicy-2"
sleep "$INITIAL_DELAY"

echo "Applying policy-2"
kubectl apply -f - <<EOF
apiVersion: policy.open-cluster-management.io/v1
kind: Policy
metadata:
  name: mypolicy-2
  namespace: $POLICY_APPLY_NS
spec:
  remediationAction: inform
  disabled: false
  policy-templates:
    - objectDefinition:
        apiVersion: policy.open-cluster-management.io/v1
        kind: ConfigurationPolicy
        metadata:
          name: policy-namespace-oscillator-2
        spec:
          remediationAction: inform
          severity: low
          namespaceSelector:
            exclude:
              - kube-*
            include:
              - default
          object-templates:
            - complianceType: $OSC_TWO
              objectDefinition:
                kind: Namespace
                apiVersion: v1
                metadata:
                  name: nonexistentnamespace
---
apiVersion: policy.open-cluster-management.io/v1
kind: PlacementBinding
metadata:
  name: binding-policy-namespace-oscillator-2
  namespace: $POLICY_APPLY_NS
placementRef:
  name: placement-policy-namespace-oscillator
  kind: PlacementRule
  apiGroup: apps.open-cluster-management.io
subjects:
  - name: mypolicy-2
    kind: Policy
    apiGroup: policy.open-cluster-management.io
EOF

CYCLE_COUNT=0
while [[ "$CYCLE_COUNT" -lt "$TOTAL_CYCLES" ]]; do
    CYCLE_COUNT=$((CYCLE_COUNT+1))
    echo "Sleep cycle $CYCLE_COUNT of $TOTAL_CYCLES"
    sleep "$SLEEP_PERIOD"

    if [[ "$OSC_ONE" == "musthave" ]]; then
        OSC_ONE="mustnothave"
    else
        OSC_ONE="musthave"
    fi
    if [[ "$OSC_TWO" == "musthave" ]]; then
        OSC_TWO="mustnothave"
    else
        OSC_TWO="musthave"
    fi

    kubectl get policy -n $POLICY_APPLY_NS
    echo "Patching mypolicy-1 to be $OSC_ONE"
    kubectl patch policy mypolicy-1 -n $POLICY_APPLY_NS --type='json' -p='[{"op": "replace", "path": "/spec/policy-templates/0/objectDefinition/spec/object-templates/0/complianceType", "value":"'$OSC_ONE'"}]'
    echo "Patching mypolicy-2 to be $OSC_TWO"
    kubectl patch policy mypolicy-2 -n $POLICY_APPLY_NS --type='json' -p='[{"op": "replace", "path": "/spec/policy-templates/0/objectDefinition/spec/object-templates/0/complianceType", "value":"'$OSC_TWO'"}]'
done

echo "Oscillations complete."
