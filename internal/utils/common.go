package utils

import (
	"math"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	danav1 "github.com/dana-team/hns/api/v1"
)

func GetSubnamespaceFromList(name string, subnamespacelist danav1.SubnamespaceList) *danav1.Subnamespace {
	for _, sns := range subnamespacelist.Items {
		if sns.Name == name {
			return &sns
		}
	}
	return nil
}

func HoursPassedSinceDate(timestamp metav1.Time) int {
	currentTime := time.Now()
	timeDiff := currentTime.Sub(timestamp.Time)
	hoursPassed := timeDiff.Hours()
	return int(math.Round(hoursPassed))
}
