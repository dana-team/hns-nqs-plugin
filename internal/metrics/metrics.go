package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var resourceOverCommitMultiplier = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "nqs_resource_over_commit_multiplier",
		Help: "Multiplier to apply to total memory in cluster",
	}, []string{"resource", "root_namespace", "secondary_root_namespace"})

// InitializeNQSMetrics initializes the metrics for NQS.
func InitializeNQSMetrics() {
	metrics.Registry.MustRegister(
		resourceOverCommitMultiplier,
	)
}

// ObserveOverCommitMultiplier sets the overcommit multiplier for a given resource.
func ObserveOverCommitMultiplier(resource, rootNS, secondaryRoot string, value float64) {
	resourceOverCommitMultiplier.With(prometheus.Labels{
		"resource":                 resource,
		"root_namespace":           rootNS,
		"secondary_root_namespace": secondaryRoot,
	}).Set(value)
}
