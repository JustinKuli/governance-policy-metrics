module github.com/JustinKuli/governance-policy-metrics

go 1.16

require (
	github.com/onsi/ginkgo v1.14.1
	github.com/onsi/gomega v1.10.2
	github.com/open-cluster-management/governance-policy-propagator v0.0.0-20210617123451-284d7175be05
	github.com/prometheus/client_golang v1.7.1 // indirect
	k8s.io/apimachinery v0.20.5
	k8s.io/client-go v12.0.0+incompatible
	sigs.k8s.io/controller-runtime v0.8.3
)

replace k8s.io/client-go => k8s.io/client-go v0.20.5
