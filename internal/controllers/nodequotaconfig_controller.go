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
	"reflect"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	danav1 "github.com/dana-team/hns/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	danav1alpha1 "nodeQuotaSync/api/v1alpha1"
	utils "nodeQuotaSync/internal/utils"
)

// NodeQuotaConfigReconciler reconciles a NodeQuotaConfig object
type NodeQuotaConfigReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=dana.hns.io,resources=nodequotaconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=v1,resources=namespace,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=dana.hns.io,resources=nodequotaconfigs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=dana.hns.io,resources=nodequotaconfigs/finalizers,verbs=update

func (r *NodeQuotaConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.Info("Start reconcile flow")
	config := danav1alpha1.NodeQuotaConfig{}
	if err := r.Get(ctx, types.NamespacedName{Name: req.Name, Namespace: req.Namespace}, &config); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}
	oldConfigStatus := config.Status.DeepCopy()
	snsList := danav1.SubnamespaceList{}
	if err := r.List(ctx, &snsList); err != nil {
		return ctrl.Result{}, err
	}
	if err, rootResourceList := r.CalculateNodeGroups(ctx, &config, snsList); err != nil {
		return ctrl.Result{}, err
	}
	utils.DeleteExpiredReservedResources(&config)
	if reflect.DeepEqual(oldConfigStatus, config.Status) {
		return ctrl.Result{}, nil
	}
	if err := r.Client.Status().Update(ctx, &config); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *NodeQuotaConfigReconciler) findConfig(node client.Object) []reconcile.Request {
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

func (r *NodeQuotaConfigReconciler) CalculateNodeGroups(ctx context.Context, config *danav1alpha1.NodeQuotaConfig, snsList danav1.SubnamespaceList) (error, v1.ResourceList) {
	rootResourceList := v1.ResourceList{}
	for _, nodegroup := range config.Spec.NodeGroupList {
		labelSelector := labels.SelectorFromSet(labels.Set(nodegroup.LabelSelector))
		listOptions := &client.ListOptions{
			LabelSelector: labelSelector,
		}
		nodeList := v1.NodeList{}
		if err := r.List(ctx, &nodeList, listOptions); err != nil {
			return err, rootResourceList
		}
		nodeResources := utils.CalculateNodeGroup(ctx, v1.NodeList{}, *config, nodegroup.Name)
		groupReserved := utils.CaculateGroupReservedResources(config.Status.ReservedResources, nodegroup.Name)
		snsObject := utils.GetSubnamespaceFromList(nodegroup.Name, snsList)
		totalResources := utils.MergeTwoResourceList(nodeResources, groupReserved)
		resourcesDiff := utils.SubstractTwoResourceList(totalResources, snsObject.Spec.ResourceQuotaSpec.Hard)
		plus, debt := utils.GetPlusAndDebtResourceList(resourcesDiff)
		config.Status.ReservedResources = []danav1alpha1.ReservedResources{
			{
				Resources: debt,
				NodeGroup: nodegroup.Name,
				Timestamp: metav1.Now(),
			},
		}
		rootResourceList = utils.MergeTwoResourceList(rootResourceList, plus)
		if !nodegroup.IsRoot {
			snsObject.Spec.ResourceQuotaSpec.Hard = utils.MergeTwoResourceList(plus, snsObject.Spec.ResourceQuotaSpec.Hard)
			if err := r.Client.Update(ctx, snsObject); err != nil {
				return err, rootResourceList
			}
		}
	}
	return nil, rootResourceList
}

// SetupWithManager sets up the controller with the Manager.
func (r *NodeQuotaConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&danav1alpha1.NodeQuotaConfig{}).
		Watches(
			&source.Kind{Type: &corev1.Node{}},
			handler.EnqueueRequestsFromMapFunc(r.findConfig),
		).
		Complete(r)
}
