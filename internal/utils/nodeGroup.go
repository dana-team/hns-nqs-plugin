package utils

import (
	"context"
	"reflect"

	danav1 "github.com/dana-team/hns/api/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	danav1alpha1 "nodeQuotaSync/api/v1alpha1"
)

func CalculateNodeGroup(ctx context.Context, nodes v1.NodeList, config danav1alpha1.NodeQuotaConfig, nodeGroup string) v1.ResourceList {
	var ResourceMultiplier map[string]string
	for _, secondaryRoot := range config.Spec.Roots {
		for _, resourceGroup := range secondaryRoot.SecondaryRoots {
			if resourceGroup.Name == nodeGroup {
				ResourceMultiplier = resourceGroup.ResourceMultiplier
			}
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

func CaculateGroupReservedResources(reserved []danav1alpha1.ReservedResources, group string, hoursLimit int) v1.ResourceList {
	resources := v1.ResourceList{}
	for _, resource := range reserved {
		if resource.NodeGroup == group && HoursPassedSinceDate(resource.Timestamp) < hoursLimit {
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

func CalculateSecondaryRootPlusDebt(ctx context.Context, r client.Client, sns danav1.Subnamespace, nodegroup danav1alpha1.NodeGroup, config *danav1alpha1.NodeQuotaConfig) (error, v1.ResourceList, v1.ResourceList) {
	labelSelector := labels.SelectorFromSet(labels.Set(nodegroup.LabelSelector))
	listOptions := &client.ListOptions{
		LabelSelector: labelSelector,
	}
	nodeList := v1.NodeList{}
	if err := r.List(ctx, &nodeList, listOptions); err != nil {
		return err, v1.ResourceList{}, v1.ResourceList{}
	}
	nodeResources := CalculateNodeGroup(ctx, nodeList, *config, nodegroup.Name)
	groupReserved := CaculateGroupReservedResources(config.Status.ReservedResources, nodegroup.Name, config.Spec.ReservedHoursTolive)
	totalResources := MergeTwoResourceList(nodeResources, groupReserved)
	resourcesDiff := SubstractTwoResourceList(totalResources, sns.Spec.ResourceQuotaSpec.Hard)
	plus, debt := GetPlusAndDebtResourceList(resourcesDiff)
	return nil, plus, debt
}

func IsReservedResourceExist(config danav1alpha1.NodeQuotaConfig, debt v1.ResourceList, nodeGroupName string) bool {
	for _, reservedResources := range config.Status.ReservedResources {
		if reservedResources.NodeGroup == nodeGroupName && reflect.DeepEqual(debt, reservedResources.Resources) {
			return true
		}
	}
	return false
}

func AddReservedToConfig(debt v1.ResourceList, nodeGroupName string, config *danav1alpha1.NodeQuotaConfig) {
	if IsReservedResourceExist(*config, debt, nodeGroupName) {
		config.Status.ReservedResources = append(config.Status.ReservedResources, danav1alpha1.ReservedResources{
			NodeGroup: nodeGroupName,
			Resources: debt,
			Timestamp: metav1.Now(),
		})
	}
}

func ProcessSecondaryRoot(ctx context.Context, r client.Client, secondaryRoot danav1alpha1.NodeGroup, config *danav1alpha1.NodeQuotaConfig, rootSubnamespace string) (error, v1.ResourceList) {
	sns := danav1.Subnamespace{}
	if err := r.Get(ctx, types.NamespacedName{Namespace: rootSubnamespace, Name: secondaryRoot.Name}, &sns); err != nil {
		return err, v1.ResourceList{}
	}
	err, plus, debt := CalculateSecondaryRootPlusDebt(ctx, r, sns, secondaryRoot, config)
	if err != nil {
		return err, v1.ResourceList{}
	}
	AddReservedToConfig(debt, secondaryRoot.Name, config)
	sns.Spec.ResourceQuotaSpec.Hard = FillterUncontrolledResources(plus, config.Spec.ControlledResources)
	if err := r.Update(ctx, &sns); err != nil {
		return err, v1.ResourceList{}
	}
	return nil, plus
}
