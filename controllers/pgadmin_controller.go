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
	pgadminv1alpha1 "github.com/dhope-nagesh/pgadmin-operator/api/v1alpha1"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// PgadminReconciler reconciles a Pgadmin object
type PgadminReconciler struct {
	client.Client
	Logger logr.Logger
	Scheme *runtime.Scheme
}

const PgAdminFinalizer string = "pgadmin.dhope-nagesh.io/finalizer"

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
	Pgadmin := &pgadminv1alpha1.Pgadmin{}

	if err := r.Get(ctx, req.NamespacedName, Pgadmin); err != nil {

		if apierrs.IsNotFound(err) {
			r.Logger.Info("Pgadmin resource not found")
			// Ignore it since its deleted
			return ctrl.Result{}, nil
		}
		r.Logger.Error(err, "Failed to fetch a pgadmin resource")
		return ctrl.Result{}, err
	}

	if err := r.ValidateCredsSecret(ctx, Pgadmin); err != nil {
		return ctrl.Result{}, err
	}

	if Pgadmin.ObjectMeta.DeletionTimestamp.IsZero() {
		if !containsString(Pgadmin.GetFinalizers(), PgAdminFinalizer) {
			controllerutil.AddFinalizer(Pgadmin, PgAdminFinalizer)
			if err := r.Update(ctx, Pgadmin); err != nil {
				return ctrl.Result{}, err
			}
			r.Logger.Info("Added Finalizer")
		}
	} else {
		r.Logger.Info("Non Zero, being deleted")
		if containsString(Pgadmin.GetFinalizers(), PgAdminFinalizer) {
			r.Logger.Info("Cleaning up")

			// Logic to clean up resources
			if err := r.DeleteDeployment(ctx, Pgadmin); err != nil {
				return ctrl.Result{}, err
			}

			controllerutil.RemoveFinalizer(Pgadmin, PgAdminFinalizer)
			if err := r.Update(ctx, Pgadmin); err != nil {
				r.Logger.Info("Cleaned up, ready to delete now...")
				return ctrl.Result{}, nil
			}
		}
		return ctrl.Result{}, nil
	}

	if err := r.CreateUpdateDeployment(ctx, Pgadmin); err != nil {
		return ctrl.Result{}, nil
	}

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

func Int32ToPtr(i int32) *int32 {
	return &i
}

func (r *PgadminReconciler) ValidateCredsSecret(ctx context.Context, Pgadmin *pgadminv1alpha1.Pgadmin) error {
	secret := &v1.Secret{}
	if err := r.Client.Get(ctx, client.ObjectKey{
		Name:      Pgadmin.Spec.CredsSecretName,
		Namespace: Pgadmin.Namespace,
	}, secret); err != nil {
		return err
	}
	return nil
}

// CreateUpdateDeployment creates or update pgadmin deployment
func (r *PgadminReconciler) CreateUpdateDeployment(ctx context.Context, Pgadmin *pgadminv1alpha1.Pgadmin) error {
	obj := appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "deployment-" + Pgadmin.Name,
			Namespace: Pgadmin.Namespace,
			Labels: map[string]string{
				"app": "pgadmin",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: Int32ToPtr(Pgadmin.Spec.Replicas),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "pgadmin",
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "pgadmin",
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  "pgadmin",
							Image: "dpage/pgadmin4",
							Ports: []v1.ContainerPort{
								{
									Name:          "http",
									ContainerPort: 80,
									Protocol:      "TCP",
								},
							},
							EnvFrom: []v1.EnvFromSource{
								{
									SecretRef: &v1.SecretEnvSource{
										LocalObjectReference: v1.LocalObjectReference{
											Name: Pgadmin.Spec.CredsSecretName,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	if err := r.Create(ctx, &obj); err != nil {
		if apierrs.IsAlreadyExists(err) {
			if err := r.Update(ctx, &obj); err != nil {
				return err
			}
			return nil
		}
		return err
	}
	return nil
}

// DeleteDeployment cleanups pgadmin deployment
func (r *PgadminReconciler) DeleteDeployment(ctx context.Context, Pgadmin *pgadminv1alpha1.Pgadmin) error {
	obj := appsv1.Deployment{}
	r.Logger.Info("Cleaning up deployment" + "deployment-" + Pgadmin.Name)

	if err := r.Client.Get(ctx, client.ObjectKey{Namespace: Pgadmin.Namespace, Name: "deployment-" + Pgadmin.Name}, &obj); err != nil {
		if apierrs.IsNotFound(err) {
			r.Logger.Info("deployment" + "deployment-" + Pgadmin.Name + "not found")
			return nil
		}
		r.Logger.Error(err, "Failed to get pgadmin deployment")
		return err
	}

	if err := r.Client.Delete(ctx, &obj); err != nil {
		r.Logger.Error(err, "Failed to delete pgadmin deployment")
		return err
	}
	return nil
}

// Helper functions to check and remove string from a slice of strings.
func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}
