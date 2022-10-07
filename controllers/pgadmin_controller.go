/*
Copyright 2022.

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
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/log"

	pgadminv1alpha1 "github.com/dhope-nagesh/pgadmin-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// PgadminReconciler reconciles a Pgadmin object
type PgadminReconciler struct {
	client.Client
	Logger logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=pgadmin.dhope-nagesh.io,resources=pgadmins,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=pgadmin.dhope-nagesh.io,resources=pgadmins/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=pgadmin.dhope-nagesh.io,resources=pgadmins/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Pgadmin object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *PgadminReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.SetLogger(ctx)

	r.Logger.Info("Getting reconciled...")

	return ctrl.Result{}, nil
}

// SetLogger set Logger object on reconciler object
func (r *PgadminReconciler) SetLogger(ctx context.Context) {
	r.Logger = log.FromContext(ctx)
}

// SetupWithManager sets up the controller with the Manager.
func (r *PgadminReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&pgadminv1alpha1.Pgadmin{}).
		Complete(r)
}
