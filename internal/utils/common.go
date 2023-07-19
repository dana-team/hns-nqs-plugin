package utils

import (
	"context"
	"math"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	danav1 "github.com/dana-team/hns/api/v1"
)

const danaHnsRoleKey = "dana.hns.io/role"

// GetSubnamespaceFromList retrieves the subnamespace with the specified name from the given subnamespace list.
// It returns a pointer to the subnamespace if found, otherwise it returns nil.
func GetSubnamespaceFromList(name string, subnamespacelist danav1.SubnamespaceList) *danav1.Subnamespace {
	for _, sns := range subnamespacelist.Items {
		if sns.Name == name {
			return &sns
		}
	}
	return nil
}

// GetRootQuota retrieves the resource quota for the root namespace with the specified name.
// It returns the root resource quota and any error encountered during retrieval.
func GetRootQuota(r client.Client, ctx context.Context, root string) (v1.ResourceQuota, error) {
	rootQuota := v1.ResourceQuota{}
	if err := r.Get(ctx, types.NamespacedName{Namespace: root, Name: root}, &rootQuota); err != nil {
		return rootQuota, err
	}
	return rootQuota, nil
}

// HoursPassedSinceDate calculates the number of hours passed since the specified timestamp.
// It returns the rounded number of hours passed.
func hoursPassedSinceDate(timestamp metav1.Time) int {
	currentTime := time.Now()
	timeDiff := currentTime.Sub(timestamp.Time)
	hoursPassed := timeDiff.Hours()
	return int(math.Round(hoursPassed))
}
