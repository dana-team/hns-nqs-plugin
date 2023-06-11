package utils

import (
	"context"

	v1 "k8s.io/api/core/v1"

	danav1alpha1 "nodeQuotaSync/api/v1alpha1"
)

func CalculateNodeGroup(ctx context.Context, nodes v1.NodeList, config danav1alpha1.NodeQuotaConfig, nodeGroup string) v1.ResourceList {
	var ResourceMultiplier map[string]string
	for _, resourceGroup := range config.Spec.NodeGroupList {
		if resourceGroup.Name == nodeGroup {
			ResourceMultiplier = resourceGroup.ResourceMultiplier
		}
	}
	nodeGroupReources := v1.ResourceList{}
	for _, node := range nodes.Items {
		resources := MultiplyResourceList(node.Status.Allocatable, ResourceMultiplier)
		for resourceName, resourceQuantity := range resources {
			AddResourcesToList(&nodeGroupReources, resourceQuantity, string(resourceName))
		}
	}

	return nodeGroupReources
}

func CaculateGroupReservedResources(reserved []danav1alpha1.ReservedResources, group string) v1.ResourceList {
	resources := v1.ResourceList{}
	for _, resource := range reserved {
		if resource.NodeGroup == group {
			for resourceName, resourceQuantity := range resources {
				AddResourcesToList(&resources, resourceQuantity, string(resourceName))
			}
		}
	}
	return resources
}

func DeleteExpiredReservedResources(config *danav1alpha1.NodeQuotaConfig) {
	newReservedResources := []danav1alpha1.ReservedResources{}
	for _, resources := range config.Status.ReservedResources {
		if HoursPassedSinceDate(resources.Timestamp) < int(config.Spec.ReservedHoursTolive) {
			newReservedResources = append(newReservedResources, resources)
		}
	}
	config.Status.ReservedResources = newReservedResources
}
