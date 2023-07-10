package utils

import (
	"context"
	"fmt"
	"reflect"

	danav1 "github.com/dana-team/hns/api/v1"
	"github.com/go-logr/logr"
	"golang.org/x/exp/slices"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	danav1alpha1 "nodeQuotaSync/api/v1alpha1"
)

// CalculateNodeGroup calculates the resource list for a node group based on the provided nodes, NodeQuotaConfig, and node group name.
// It takes a context, a NodeList containing the nodes, the NodeQuotaConfig, and the node group name.
// It returns the calculated resource list (v1.ResourceList) for the node group.
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
		resources := multiplyResourceList(node.Status.Allocatable, ResourceMultiplier)
		for resourceName, resourceQuantity := range resources {
			addResourcesToList(&nodeGroupReources, resourceQuantity, string(resourceName))
		}
	}

	return filterUncontrolledResources(nodeGroupReources, config.Spec.ControlledResources)
}

// getReservedResourcesByGroup retrieves the reserved resources for a specific node group from the NodeQuotaConfig.
// It takes the node group name and the NodeQuotaConfig.
// It returns the ReservedResources object (danav1alpha1.ReservedResources) for the node group.
func getReservedResourcesByGroup(group string, config danav1alpha1.NodeQuotaConfig) danav1alpha1.ReservedResources {
	if !doesReservedResourceExist(config, group) {
		return danav1alpha1.ReservedResources{}
	}
	for _, resource := range config.Status.ReservedResources {
		if resource.NodeGroup == group {
			return resource
		}
	}
	return danav1alpha1.ReservedResources{}
}

// DeleteExpiredReservedResources removes the expired reserved resources from the NodeQuotaConfig.
// It takes the NodeQuotaConfig to modify and a logger for logging informational messages.
func DeleteExpiredReservedResources(config *danav1alpha1.NodeQuotaConfig, logger logr.Logger) {
	newReservedResources := []danav1alpha1.ReservedResources{}
	for _, resources := range config.Status.ReservedResources {
		if isReservedResourceExpired(resources, *config) {
			logger.Info(fmt.Sprintf("Removed ReservedResources from nodeGroup %s", resources.NodeGroup))
		} else {
			newReservedResources = append(newReservedResources, resources)
		}
	}
	config.Status.ReservedResources = newReservedResources
}

// CalculateSecondaryNodeGroup calculates the resource list for a secondary node group based on the provided nodegroup and NodeQuotaConfig.
// It takes a context, a client for making API requests, a nodegroup to calculate resources for, and the NodeQuotaConfig.
// It returns an error (if any occurred) and the calculated resource list (v1.ResourceList).
func CalculateSecondaryNodeGroup(ctx context.Context, r client.Client, nodegroup danav1alpha1.NodeGroup, config *danav1alpha1.NodeQuotaConfig) (error, v1.ResourceList) {
	logr, _ := logr.FromContext(ctx)
	labelSelector := labels.SelectorFromSet(labels.Set(nodegroup.LabelSelector))
	listOptions := &client.ListOptions{
		LabelSelector: labelSelector,
	}
	nodeList := v1.NodeList{}
	if err := r.List(ctx, &nodeList, listOptions); err != nil {
		logr.Error(err, fmt.Sprintf("Error listing the nodes for the nodeGroup %s", nodegroup))
		return err, v1.ResourceList{}
	}
	nodeResources := CalculateNodeGroup(ctx, nodeList, *config, nodegroup.Name)
	return nil, nodeResources
}

// doesReservedResourceExist checks if a reserved resource exists in the NodeQuotaConfig for the given node group name.
// It takes the NodeQuotaConfig and the node group name to check.
// It returns a boolean value indicating whether the reserved resource exists or not.
func doesReservedResourceExist(config danav1alpha1.NodeQuotaConfig, nodeGroupName string) bool {
	if len(config.Status.ReservedResources) == 0 {
		return false
	}
	for _, reservedResources := range config.Status.ReservedResources {
		if reservedResources.NodeGroup == nodeGroupName {
			return true
		}
	}
	return false
}

// UpdateRootSubnamespace updates the resourceQuota of the rootSubnamespace with the new quantity of resources.
func UpdateRootSubnamespace(ctx context.Context, rootResources v1.ResourceList, rootSubnamespace danav1alpha1.SubnamespacesRoots, logger logr.Logger, client client.Client) error {
	rootRQ, err := GetRootQuota(client, ctx, rootSubnamespace.RootNamespace)
	if err != nil {
		logger.Error(err, fmt.Sprintf("Error getting the %s resourceQuota", rootSubnamespace.RootNamespace))
		return err
	}
	rootRQ.Spec.Hard = rootResources
	logger.Info(fmt.Sprintf("Updating RootSubnamespace %s with new resources", rootSubnamespace.RootNamespace))
	if err := client.Update(ctx, &rootRQ); err != nil {
		logger.Error(err, fmt.Sprintf("Error updating rootSubnamespace %s"), rootSubnamespace.RootNamespace)
		return err
	}
	return nil
}

// UpdateRootSubnamespace updates the secondaryRoots in the cluster with the new quantity of resources.
// It takes slice of Subnamespaces that was updated in memory and does API requests to commit the update.
func UpdateProcessedSecondaryRoots(ctx context.Context, processedSecondaryRoots []danav1.Subnamespace, logger logr.Logger, client client.Client) error {
	for _, sns := range processedSecondaryRoots {
		logger.Info(fmt.Sprintf("Updating secondaryRoot %s with new resources", sns.Name))
		if err := client.Update(ctx, &sns); err != nil {
			logger.Error(err, fmt.Sprintf("Error updating secondaryRoot %s"), sns.Name)
			return err
		}
	}
	return nil
}

// isReservedResourceExpired checks if a reserved created more than X hours ago, defined by the user in the config CRD.
func isReservedResourceExpired(reservedResources danav1alpha1.ReservedResources, config danav1alpha1.NodeQuotaConfig) bool {
	return hoursPassedSinceDate(reservedResources.Timestamp) >= config.Spec.ReservedHoursToLive
}

// setReservedToConfig sets the reserved resources for a node group in the NodeQuotaConfig.
// It takes the resource debt (v1.ResourceList) to set, the node group name, the NodeQuotaConfig to modify, and a logger for logging informational messages.
func setReservedToConfig(debt v1.ResourceList, nodeGroupName string, config *danav1alpha1.NodeQuotaConfig, logr logr.Logger) {
	if doesReservedResourceExist(*config, nodeGroupName) {
		reservedResources := getReservedResourcesByGroup(nodeGroupName, *config)
		reservedResources.Resources = debt
		removeReservedFromConfig(nodeGroupName, config)
		config.Status.ReservedResources = append(config.Status.ReservedResources, reservedResources)
		return
	}
	config.Status.ReservedResources = append(config.Status.ReservedResources, danav1alpha1.ReservedResources{
		NodeGroup: nodeGroupName,
		Resources: debt,
		Timestamp: metav1.Now(),
	})
	logr.Info(fmt.Sprintf("Added ReservedResources to nodeGroup %s", nodeGroupName))
}

// removeReservedFromConfig removes the reserved resources for a node group from the NodeQuotaConfig.
// It takes the node group name and the NodeQuotaConfig to modify.
func removeReservedFromConfig(nodeGroupName string, config *danav1alpha1.NodeQuotaConfig) {
	index := -1
	for i, reservedResources := range config.Status.ReservedResources {
		if reservedResources.NodeGroup == nodeGroupName {
			index = i
		}
	}
	if index == -1 {
		return
	}
	config.Status.ReservedResources = slices.Delete(config.Status.ReservedResources, index, index+1)
}

// ProcessSecondaryRoot processes a secondary root node group and updates the corresponding Subnamespace object and add reserved resources to the config if needed.
// It takes a context, a client for making API requests, the secondary root node group, the NodeQuotaConfig,
// the root subnamespace, and a logger for logging informational messages.
// It returns an error (if any occurred) and the updated Subnamespace object (danav1.Subnamespace).
func ProcessSecondaryRoot(ctx context.Context, r client.Client, secondaryRoot danav1alpha1.NodeGroup, config *danav1alpha1.NodeQuotaConfig, rootSubnamespace string, logr logr.Logger) (error, danav1.Subnamespace) {
	sns := danav1.Subnamespace{}
	if err := r.Get(ctx, types.NamespacedName{Namespace: rootSubnamespace, Name: secondaryRoot.Name}, &sns); err != nil {
		logr.Error(err, fmt.Sprintf("Error getting the subnamespace %s", secondaryRoot.Name))
		return err, sns
	}
	err, groupResources := CalculateSecondaryNodeGroup(ctx, r, secondaryRoot, config)
	if err != nil {
		return err, sns
	}
	groupReserved := getReservedResourcesByGroup(secondaryRoot.Name, *config)
	if isGreaterThan(sns.Spec.ResourceQuotaSpec.Hard, groupResources) {
		// Nodes removed
		debt := subtractTwoResourceList(sns.Spec.ResourceQuotaSpec.Hard, groupResources)
		if groupReserved.NodeGroup == "" || !isReservedResourceExpired(groupReserved, *config) {
			setReservedToConfig(debt, secondaryRoot.Name, config, logr)
			return nil, sns
		}
	} else {
		// Nodes added
		totalResources := MergeTwoResourceList(groupResources, groupReserved.Resources)
		if isGreaterThan(totalResources, sns.Spec.ResourceQuotaSpec.Hard) || isEqualTo(totalResources, sns.Spec.ResourceQuotaSpec.Hard) {
			removeReservedFromConfig(secondaryRoot.Name, config)
		}
	}
	if !reflect.DeepEqual(sns.Spec.ResourceQuotaSpec.Hard, groupResources) {
		sns.Spec.ResourceQuotaSpec.Hard = filterUncontrolledResources(groupResources, config.Spec.ControlledResources)
	}
	return nil, sns
}
