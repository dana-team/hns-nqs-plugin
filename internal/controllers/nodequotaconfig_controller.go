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
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
	_ = log.FromContext(ctx)
	config := danav1alpha1.NodeQuotaConfig{}
	if err := r.Get(ctx, types.NamespacedName{Name: req.Name, Namespace: req.Namespace}, &config); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}
	if err := r.CalculateRootSubnamespaces(ctx, config); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NodeQuotaConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&danav1alpha1.NodeQuotaConfig{}).
		Watches(
			&source.Kind{Type: &corev1.Node{}},
			handler.EnqueueRequestsFromMapFunc(r.requestConfigReconcile),
		).
		Complete(r)
}

func (r *NodeQuotaConfigReconciler) CalculateRootSubnamespaces(ctx context.Context, config danav1alpha1.NodeQuotaConfig) error {
	oldConfigStatus := config.Status.DeepCopy()
	reservedResources := []danav1alpha1.ReservedResources{}
	for _, rootSubnamespace := range config.Spec.Roots {
		rootResources := v1.ResourceList{}
		for _, secondaryRoot := range rootSubnamespace.SecondaryRoots {
			err, secondaryRootResources := utils.ProcessSecondaryRoot(ctx, r.Client, secondaryRoot, &config, rootSubnamespace.RootNamespace)
			if err != nil {
				return err
			}
			rootResources = utils.MergeTwoResourceList(secondaryRootResources, rootResources)
		}
		rootRQ, err := utils.GetRootQuota(r.Client, ctx, rootSubnamespace.RootNamespace)
		if err != nil {
			return err
		} else {
			rootRQ.Spec.Hard = utils.FillterUncontrolledResources(rootResources, config.Spec.ControlledResources)
			if err := r.Update(ctx, &rootRQ); err != nil {
				return err
			}
		}
	}
	config.Status.ReservedResources = reservedResources
	utils.DeleteExpiredReservedResources(&config)
	if !reflect.DeepEqual(oldConfigStatus, config.Status) {
		if err := r.Status().Update(ctx, &config); err != nil {
			return err
		}
	}
	return nil
}

func (r *NodeQuotaConfigReconciler) requestConfigReconcile(node client.Object) []reconcile.Request {
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
