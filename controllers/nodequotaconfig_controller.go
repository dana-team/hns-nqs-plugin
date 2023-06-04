/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"time"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	danav1alpha1 "nodeQuotaSync/api/v1alpha1"
)

// NodeQuotaConfigReconciler reconciles a NodeQuotaConfig object
type NodeQuotaConfigReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=dana.hns.io,resources=nodequotaconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=dana.hns.io,resources=nodequotaconfigs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=dana.hns.io,resources=nodequotaconfigs/finalizers,verbs=update

func (r *NodeQuotaConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.Info("Start reconcile file")
	nodeQuotaConfig := danav1alpha1.NodeQuotaConfigList{}

	if err := r.Client.List(ctx, &nodeQuotaConfig); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func MultiplyResourceList(resources v1.ResourceList, factor map[string]float64) v1.ResourceList {
	result := make(v1.ResourceList)

	for name, value := range resources {
		newValue := new(resource.Quantity)
		newValue.Set(value.Value())

		// Multiply the value by the factor
		newValueInt64 := float64(newValue.Value()) * factor[name.String()]
		newValue.Set(int64(newValueInt64))

		result[name] = *newValue
	}

	return result
}

func ClaculateNodeWithMultiplier(node v1.Node, multiplier danav1alpha1.NodeGroup) v1.ResourceList {
	resourceList := v1.ResourceList{}

	return resourceList
}

func HoursPassedSinceDate(timestamp metav1.Time) float64 {
	currentTime := time.Now()
	timeDiff := currentTime.Sub(timestamp.Time)
	hoursPassed := timeDiff.Hours()
	return hoursPassed
}

func CalculateNodeGroup(ctx context.Context, nodes v1.NodeList, config danav1alpha1.NodeQuotaConfig, nodeGroup string) (v1.ResourceList, v1.ResourceList) {
	calculateReserved := make(v1.ResourceList)
	var ResourceMultiplier map[string]float64
	for _, resourceGroup := range config.Spec.NodeGroupList {
		if resourceGroup.Name == nodeGroup {
			ResourceMultiplier = resourceGroup.ResourceMultiplier
		}
	}
	for name, row := range config.Status.ReservedResources[nodeGroup] {

	}
	debtResources := v1.ResourceList{}
	nodeGroupReources := v1.ResourceList{}
	return debtResources, nodeGroupReources

}

func (r *NodeQuotaConfigReconciler) findNodes(node client.Object) []reconcile.Request {
	nodeQuotaConfig := danav1alpha1.NodeQuotaConfigList{}
	err := r.List(context.TODO(), &nodeQuotaConfig)
	if err != nil {
		return []reconcile.Request{}
	}

	requests := make([]reconcile.Request, len(nodeQuotaConfig.Items))
	for i, item := range nodeQuotaConfig.Items {
		requests[i] = reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      item.GetName(),
				Namespace: item.GetNamespace(),
			},
		}
	}
	return requests
}

// SetupWithManager sets up the controller with the Manager.
func (r *NodeQuotaConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&danav1alpha1.NodeQuotaConfig{}).
		Watches(
			&source.Kind{Type: &corev1.Node{}},
			handler.EnqueueRequestsFromMapFunc(r.findNodes),
		).
		Complete(r)
}
