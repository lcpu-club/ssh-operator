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

package v1

import (
	"context"
	"fmt"
	"strings"

	_ "embed"

	sshoperatorv1alpha1 "github.com/lcpu-club/ssh-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// nolint:unused
// log is for logging in this package.
var podlog = logf.Log.WithName("pod-resource")

// SetupPodWebhookWithManager registers the webhook for Pod in the manager.
func SetupPodWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&corev1.Pod{}).
		WithDefaulter(&PodCustomDefaulter{
			c: mgr.GetClient(),
		}).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// +kubebuilder:webhook:path=/mutate--v1-pod,mutating=true,failurePolicy=fail,sideEffects=None,groups="",resources=pods,verbs=create;update,versions=v1,name=mpod-v1.kb.io,admissionReviewVersions=v1,matchPolicy=Equivalent

// PodCustomDefaulter struct is responsible for setting default values on the custom resource of the
// Kind Pod when those are created or updated.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as it is used only for temporary operations and does not need to be deeply copied.
type PodCustomDefaulter struct {
	c client.Client
}

// +kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch

var _ webhook.CustomDefaulter = &PodCustomDefaulter{}

//go:embed inject_script.sh
var injectScript string

// Default implements webhook.CustomDefaulter so a webhook will be registered for the Kind Pod.
func (d *PodCustomDefaulter) Default(ctx context.Context, obj runtime.Object) error {
	pod, ok := obj.(*corev1.Pod)

	if !ok {
		return fmt.Errorf("expected an Pod object but got %T", obj)
	}
	podlog.Info("Defaulting for Pod", "name", pod.GetName())

	// TODO(user): fill in your defaulting logic.
	ns := pod.GetNamespace()
	nsObj := &corev1.Namespace{}
	if err := d.c.Get(ctx, client.ObjectKey{Name: ns}, nsObj); err != nil {
		return fmt.Errorf("failed to get namespace: %w", err)
	}
	nsLabel, ok := nsObj.ObjectMeta.Labels["ssh-operator.lcpu.dev/inject"]
	if !ok || nsLabel != "enabled" {
		if nsLabel != "conditional" {
			return nil
		}
		if an, ok := pod.ObjectMeta.Annotations["ssh-operator.lcpu.dev/inject"]; !ok || an != "enabled" {
			return nil
		}
	}
	if an, ok := pod.ObjectMeta.Annotations["ssh-operator.lcpu.dev/inject"]; ok && an == "disabled" {
		return nil
	}

	// ssh-operator.lcpu.dev/inject-path is used to specify the path to inject the .ssh directory
	// Default: /root
	// Required: false
	injectPath := "/root"
	if path, ok := pod.ObjectMeta.Annotations["ssh-operator.lcpu.dev/inject-path"]; ok {
		injectPath = path
	}
	// ssh-operator.lcpu.dev/modify-command is used to specify whether to modify the command
	// Default: true
	// Required: false
	modifyCommand := true
	if cmd, ok := pod.ObjectMeta.Annotations["ssh-operator.lcpu.dev/modify-command"]; ok {
		modifyCommand = (cmd == "true") || (cmd == "yes") || (cmd == "on") || (cmd == "enabled")
	}

	// Avoid re-injecting the .ssh directory
	for _, volume := range pod.Spec.Volumes {
		if volume.Name == "dot-ssh" {
			return nil
		}
	}

	// Get ssh key pairs & authorized keys
	kpList := &sshoperatorv1alpha1.SSHKeyPairList{}
	if err := d.c.List(ctx, kpList, client.InNamespace(ns)); err != nil {
		return fmt.Errorf("failed to list SSHKeyPair objects: %w", err)
	}

	akList := &sshoperatorv1alpha1.SSHAuthorizedKeyList{}
	if err := d.c.List(ctx, akList, client.InNamespace(ns)); err != nil {
		return fmt.Errorf("failed to list SSHAuthorizedKey objects: %w", err)
	}

	akString := ""
	for _, ak := range akList.Items {
		if ak.Spec.Key == "" {
			continue
		}
		akString += ak.Spec.Key + "\n"
	}
	akString = strings.Trim(akString, "\n")

	kpNames := []string{}
	kpDefs := []corev1.EnvVar{}
	for _, kp := range kpList.Items {
		kpNames = append(kpNames, kp.Spec.Name)
		kpDefs = append(kpDefs, corev1.EnvVar{
			Name:  "SSH_KEY_" + kp.Spec.Name + "_PRIVATE",
			Value: kp.Spec.PrivateKey,
		}, corev1.EnvVar{
			Name:  "SSH_KEY_" + kp.Spec.Name + "_PUBLIC",
			Value: kp.Spec.PublicKey,
		})
	}

	if pod.ObjectMeta.Annotations == nil {
		pod.ObjectMeta.Annotations = make(map[string]string)
	}
	pod.ObjectMeta.Annotations["ssh-operator.lcpu.dev/authorized_keys"] = akString
	pod.Spec.Volumes = append(pod.Spec.Volumes, corev1.Volume{
		Name: "dot-ssh",
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{
				Medium: corev1.StorageMediumDefault,
			},
		},
	})
	pod.Spec.InitContainers = append(pod.Spec.InitContainers, corev1.Container{
		Name:  "init-ssh",
		Image: "alpine",
		Command: []string{
			"sh",
			"-c",
			injectScript,
		},
		Env: append([]corev1.EnvVar{
			{
				Name:  "AUTHORIZED_KEYS",
				Value: akString,
			},
			{
				Name:  "SSH_KEY_PAIRS",
				Value: strings.Join(kpNames, " "),
			},
		}, kpDefs...),
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      "dot-ssh",
				MountPath: "/mnt/.ssh",
				ReadOnly:  false,
			},
		},
	})
	for k := range pod.Spec.Containers {
		pod.Spec.Containers[k].VolumeMounts = append(pod.Spec.Containers[k].VolumeMounts, corev1.VolumeMount{
			Name:      "dot-ssh",
			MountPath: fmt.Sprintf("%s/.ssh", injectPath),
			ReadOnly:  false,
		})
	}
	if pod.Spec.Hostname == "" {
		pod.Spec.Hostname = pod.GetName()
	}

	if len(pod.Spec.Containers) < 1 {
		return nil
	}

	if modifyCommand {
		origCmd := pod.Spec.Containers[0].Command
		pod.Spec.Containers[0].Command = append([]string{
			"/bin/bash",
			"-c",
			// Use exec to replace the shell process with the user's command
			// Accept any environment variables from the client
			fmt.Sprintf("chmod -R 0600 %s/.ssh; mkdir -p /run/sshd; /sbin/sshd -o \"AcceptEnv=*\"; exec \"$@\"", injectPath),
			"--",
		}, origCmd...)
	}

	return nil
}
