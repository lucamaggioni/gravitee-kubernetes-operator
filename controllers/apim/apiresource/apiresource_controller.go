/*
 * Copyright (C) 2015 The Gravitee team (http://gravitee.io)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *         http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package apiresource

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	gio "github.com/gravitee-io/gravitee-kubernetes-operator/api/v1alpha1"
)

// Reconciler reconciles a ApiResource object.
type Reconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=gravitee.io,resources=apiresources,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=gravitee.io,resources=apiresources/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=gravitee.io,resources=apiresources/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx).WithValues("ApiResource", req.NamespacedName)

	instance := &gio.ApiResource{}

	if err := r.Get(ctx, req.NamespacedName, instance); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if instance.IsBeingDeleted() {
		return ctrl.Result{}, nil
	}

	log.Info("Reconciling ApiResource instance")

	// Update API resources that reference this resource
	apis, err := r.listApiResourcesWithReference(ctx, instance.Name, instance.Namespace)
	if err != nil {
		log.Error(err, "unable to list API definitions resources, skipping update")
		return ctrl.Result{}, nil
	}

	for i := range apis {
		if retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			api := apis[i]
			api.Status.ProcessingStatus = gio.ProcessingStatusReconciling
			log.Info("updating API definition", "api", api.Name)
			return r.Status().Update(ctx, &api)
		}); retryErr != nil {
			log.Error(retryErr, "unable to update API definition status, skipping update")
		}
	}

	return ctrl.Result{}, nil
}

func (r *Reconciler) listApiResourcesWithReference(
	ctx context.Context, resourceName, resourceNamespace string,
) ([]gio.ApiDefinition, error) {
	log := log.FromContext(ctx)

	apiDefinitionList := &gio.ApiDefinitionList{}
	results := make([]gio.ApiDefinition, 0)

	if err := r.Client.List(ctx, apiDefinitionList); err != nil {
		log.Error(err, "unable to list API definitions, skipping update")
		return nil, err
	}

	for _, api := range apiDefinitionList.Items {
		if api.Spec.Resources == nil {
			continue
		}

		for _, resource := range api.Spec.Resources {
			if resource.IsMatchingRef(resourceName, resourceNamespace) {
				results = append(results, api)
			}
		}
	}

	return results, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&gio.ApiResource{}).
		Complete(r)
}
