package utils

import (
	"context"
	"math"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	danav1 "github.com/dana-team/hns/api/v1"
)

const danaHnsRoleKey = "dana.hns.io/role"

func GetSubnamespaceFromList(name string, subnamespacelist danav1.SubnamespaceList) *danav1.Subnamespace {
	for _, sns := range subnamespacelist.Items {
		if sns.Name == name {
			return &sns
		}
	}
	return nil
}

func ContainesRootAnnotation(namespace v1.Namespace) bool {
	annos := namespace.GetAnnotations()
	if len(annos) == 0 {
		return false
	}

	in, ok := annos[danaHnsRoleKey]
	return ok && len(in) > 0
}

func GetRootQuota(namepaces v1.NamespaceList, r client.Client, ctx context.Context) (v1.ResourceQuota, error) {
	rootQuota := v1.ResourceQuota{}
	for _, namespace := range namepaces.Items {
		if ContainesRootAnnotation(namespace) {
			resourceQuota := v1.ResourceQuotaList{}
			if err := r.List(ctx, &resourceQuota); err != nil {
				return rootQuota, nil
			}
			return resourceQuota.Items[0], nil
		}
	}
	return rootQuota, nil
}

func HoursPassedSinceDate(timestamp metav1.Time) int {
	currentTime := time.Now()
	timeDiff := currentTime.Sub(timestamp.Time)
	hoursPassed := timeDiff.Hours()
	return int(math.Round(hoursPassed))
}
