/*
Copyright 2024.

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

package controller

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	sshoperatorv1alpha1 "github.com/lcpu-club/ssh-operator/api/v1alpha1"
	"github.com/lcpu-club/ssh-operator/internal/utils"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

// SSHKeyPairReconciler reconciles a SSHKeyPair object
type SSHKeyPairReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

const finalizerName = "ssh-operator.kube.lcpu.dev/sshkeypair-finalizer"

// +kubebuilder:rbac:groups=ssh-operator.lcpu.dev,resources=sshkeypairs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=ssh-operator.lcpu.dev,resources=sshkeypairs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=ssh-operator.lcpu.dev,resources=sshkeypairs/finalizers,verbs=update
// +kubebuilder:rbac:groups=ssh-operator.lcpu.dev,resources=sshauthorizedkeys,verbs=get;list;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the SSHKeyPair object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.1/pkg/reconcile
func (r *SSHKeyPairReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// TODO(user): your logic here
	keyPair := &sshoperatorv1alpha1.SSHKeyPair{}
	if err := r.Get(ctx, req.NamespacedName, keyPair); err != nil {
		log.Error(err, "unable to fetch SSHKeyPair")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	notFound := false
	authorizedKey := &sshoperatorv1alpha1.SSHAuthorizedKey{}
	if err := r.Get(ctx, req.NamespacedName, authorizedKey); err != nil {
		if apierrors.IsNotFound(err) {
			notFound = true
		} else {
			log.Error(err, "unable to fetch SSHAuthorizedKey")
			return ctrl.Result{}, err
		}
	}

	authorizedKey.SetName(req.Name)
	authorizedKey.SetNamespace(req.Namespace)
	authorizedKey.Spec.Key = keyPair.Spec.PublicKey

	if keyPair.DeletionTimestamp.IsZero() {
		if !utils.ContainsString(keyPair.GetFinalizers(), finalizerName) {
			keyPair.SetFinalizers(append(keyPair.GetFinalizers(), finalizerName))
			if err := r.Update(ctx, keyPair); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		if utils.ContainsString(keyPair.GetFinalizers(), finalizerName) {
			if err := r.Delete(ctx, authorizedKey); client.IgnoreNotFound(err) != nil {
				return ctrl.Result{}, err
			}
			keyPair.SetFinalizers(utils.RemoveString(keyPair.GetFinalizers(), finalizerName))
			if err := r.Update(ctx, keyPair); err != nil {
				return ctrl.Result{}, err
			}
		}
	}

	if notFound {
		if err := r.Create(ctx, authorizedKey); client.IgnoreAlreadyExists(err) != nil {
			log.Error(err, "unable to create SSHAuthorizedKey")
			return ctrl.Result{}, err
		}
	} else {
		if err := r.Update(ctx, authorizedKey); client.IgnoreNotFound(err) != nil {
			log.Error(err, "unable to update SSHAuthorizedKey")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SSHKeyPairReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		// Uncomment the following line adding a pointer to an instance of the controlled resource as an argument
		// For().
		Named("sshkeypair").
		For(&sshoperatorv1alpha1.SSHKeyPair{}).
		Owns(&sshoperatorv1alpha1.SSHAuthorizedKey{}).
		Complete(r)
}
