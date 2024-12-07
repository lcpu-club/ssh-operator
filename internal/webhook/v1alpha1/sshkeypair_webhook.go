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

package v1alpha1

import (
	"context"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	sshoperatorv1alpha1 "github.com/lcpu-club/ssh-operator/api/v1alpha1"
	"github.com/lcpu-club/ssh-operator/internal/utils"
)

// nolint:unused
// log is for logging in this package.
var sshkeypairlog = logf.Log.WithName("sshkeypair-resource")

// SetupSSHKeyPairWebhookWithManager registers the webhook for SSHKeyPair in the manager.
func SetupSSHKeyPairWebhookWithManager(mgr ctrl.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &sshoperatorv1alpha1.SSHKeyPair{}, ".spec.name", func(rawObj client.Object) []string {
		sshkeypair := rawObj.(*sshoperatorv1alpha1.SSHKeyPair)
		return []string{sshkeypair.Spec.Name}
	}); err != nil {
		return fmt.Errorf("failed to set index for SSHKeyPair: %w", err)
	}

	return ctrl.NewWebhookManagedBy(mgr).For(&sshoperatorv1alpha1.SSHKeyPair{}).
		WithValidator(&SSHKeyPairCustomValidator{
			c: mgr.GetClient(),
		}).
		WithDefaulter(&SSHKeyPairCustomDefaulter{
			DefaultType:       "ssh-ed25519",
			DefaultNamePrefix: "id_",
		}).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// +kubebuilder:webhook:path=/mutate-ssh-operator-lcpu-dev-v1alpha1-sshkeypair,mutating=true,failurePolicy=fail,sideEffects=None,groups=ssh-operator.lcpu.dev,resources=sshkeypairs,verbs=create;update,versions=v1alpha1,name=msshkeypair-v1alpha1.kb.io,admissionReviewVersions=v1

// SSHKeyPairCustomDefaulter struct is responsible for setting default values on the custom resource of the
// Kind SSHKeyPair when those are created or updated.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as it is used only for temporary operations and does not need to be deeply copied.
type SSHKeyPairCustomDefaulter struct {
	DefaultType       string
	DefaultNamePrefix string
}

var _ webhook.CustomDefaulter = &SSHKeyPairCustomDefaulter{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the Kind SSHKeyPair.
func (d *SSHKeyPairCustomDefaulter) Default(ctx context.Context, obj runtime.Object) error {
	sshkeypair, ok := obj.(*sshoperatorv1alpha1.SSHKeyPair)

	if !ok {
		return fmt.Errorf("expected an SSHKeyPair object but got %T", obj)
	}
	sshkeypairlog.Info("Defaulting for SSHKeyPair", "name", sshkeypair.GetName())

	if sshkeypair.Spec.Type == "" {
		sshkeypair.Spec.Type = d.DefaultType
	}

	if sshkeypair.Spec.PrivateKey == "" {
		if sshkeypair.Spec.PublicKey != "" {
			return nil
		}

		// Generate a new key pair
		pub, priv, err := utils.GenerateKeyPair(sshkeypair.Spec.Type)
		if err != nil {
			return fmt.Errorf("failed to generate key pair: %w", err)
		}

		sshkeypair.Spec.PublicKey = pub
		sshkeypair.Spec.PrivateKey = priv
	} else if sshkeypair.Spec.PublicKey == "" {
		pub, err := utils.PublicKeyFromPrivateKey(sshkeypair.Spec.PrivateKey)
		if err != nil {
			return fmt.Errorf("failed to extract public key: %w", err)
		}
		sshkeypair.Spec.PublicKey = pub
	}

	typ, err := utils.CheckKeyPair(sshkeypair.Spec.PublicKey, sshkeypair.Spec.PrivateKey)
	if err != nil {
		return nil
	}
	sshkeypair.Spec.Type = typ

	if sshkeypair.Spec.Name == "" {
		sshkeypair.Spec.Name = fmt.Sprintf("%s%s", d.DefaultNamePrefix,
			strings.TrimPrefix(sshkeypair.Spec.Type, "ssh-"),
		)
	}

	return nil
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// NOTE: The 'path' attribute must follow a specific pattern and should not be modified directly here.
// Modifying the path for an invalid path can cause API server errors; failing to locate the webhook.
// +kubebuilder:webhook:path=/validate-ssh-operator-lcpu-dev-v1alpha1-sshkeypair,mutating=false,failurePolicy=fail,sideEffects=None,groups=ssh-operator.lcpu.dev,resources=sshkeypairs,verbs=create;update,versions=v1alpha1,name=vsshkeypair-v1alpha1.kb.io,admissionReviewVersions=v1

// SSHKeyPairCustomValidator struct is responsible for validating the SSHKeyPair resource
// when it is created, updated, or deleted.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as this struct is used only for temporary operations and does not need to be deeply copied.
type SSHKeyPairCustomValidator struct {
	c client.Client
}

var _ webhook.CustomValidator = &SSHKeyPairCustomValidator{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type SSHKeyPair.
func (v *SSHKeyPairCustomValidator) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	sshkeypair, ok := obj.(*sshoperatorv1alpha1.SSHKeyPair)
	if !ok {
		return nil, fmt.Errorf("expected a SSHKeyPair object but got %T", obj)
	}
	sshkeypairlog.Info("Validation for SSHKeyPair upon creation", "name", sshkeypair.GetName())

	return nil, v.validateSSHKeyPair(ctx, sshkeypair)
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type SSHKeyPair.
func (v *SSHKeyPairCustomValidator) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	sshkeypair, ok := newObj.(*sshoperatorv1alpha1.SSHKeyPair)
	if !ok {
		return nil, fmt.Errorf("expected a SSHKeyPair object for the newObj but got %T", newObj)
	}
	sshkeypairlog.Info("Validation for SSHKeyPair upon update", "name", sshkeypair.GetName())

	if !sshkeypair.DeletionTimestamp.IsZero() {
		// Being deleting, no need to validate
		return nil, nil
	}

	return nil, v.validateSSHKeyPair(ctx, sshkeypair)
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type SSHKeyPair.
func (v *SSHKeyPairCustomValidator) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	sshkeypair, ok := obj.(*sshoperatorv1alpha1.SSHKeyPair)
	if !ok {
		return nil, fmt.Errorf("expected a SSHKeyPair object but got %T", obj)
	}
	sshkeypairlog.Info("Validation for SSHKeyPair upon deletion", "name", sshkeypair.GetName())

	return nil, nil
}

func (v *SSHKeyPairCustomValidator) validateSSHKeyPair(ctx context.Context, sshkeypair *sshoperatorv1alpha1.SSHKeyPair) error {
	if sshkeypair.Spec.PublicKey == "" {
		return fmt.Errorf("public key is missing")
	}

	if sshkeypair.Spec.PrivateKey == "" {
		return fmt.Errorf("private key is missing")
	}

	if sshkeypair.Spec.Type == "" {
		return fmt.Errorf("type is missing")
	}

	typ, err := utils.CheckKeyPair(sshkeypair.Spec.PublicKey, sshkeypair.Spec.PrivateKey)
	if err != nil {
		return err
	}

	if typ != sshkeypair.Spec.Type {
		return fmt.Errorf("type does not match the key pair")
	}

	kpl := &sshoperatorv1alpha1.SSHKeyPairList{}
	if err := v.c.List(ctx, kpl, client.MatchingFields{".spec.name": sshkeypair.Spec.Name}, client.InNamespace(sshkeypair.GetNamespace())); err != nil {
		return fmt.Errorf("failed to list SSHKeyPair resources: %w", err)
	}

	if len(kpl.Items) > 1 {
		return fmt.Errorf("duplicate SSHKeyPair resources with the same spec.name")
	}
	if len(kpl.Items) == 1 && kpl.Items[0].GetName() != sshkeypair.GetName() {
		return fmt.Errorf("duplicate SSHKeyPair resources with the same spec.name")
	}

	return nil
}
