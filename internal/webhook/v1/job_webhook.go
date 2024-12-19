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
	"strconv"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// nolint:unused
// log is for logging in this package.
var joblog = logf.Log.WithName("job-resource")

// SetupJobWebhookWithManager registers the webhook for Job in the manager.
func SetupJobWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&batchv1.Job{}).
		WithDefaulter(&JobCustomDefaulter{}).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// +kubebuilder:webhook:path=/mutate-batch-v1-job,mutating=true,failurePolicy=fail,sideEffects=None,groups=batch,resources=jobs,verbs=create;update,versions=v1,name=mjob-v1.kb.io,admissionReviewVersions=v1

// JobCustomDefaulter struct is responsible for setting default values on the custom resource of the
// Kind Job when those are created or updated.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as it is used only for temporary operations and does not need to be deeply copied.
type JobCustomDefaulter struct {
	// TODO(user): Add more fields as needed for defaulting
}

var _ webhook.CustomDefaulter = &JobCustomDefaulter{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the Kind Job.
func (d *JobCustomDefaulter) Default(ctx context.Context, obj runtime.Object) error {
	job, ok := obj.(*batchv1.Job)

	if !ok {
		return fmt.Errorf("expected an Job object but got %T", obj)
	}
	joblog.Info("Defaulting for Job", "name", job.GetName())

	// TODO(user): fill in your defaulting logic.
	// TODO: More robust logic
	if *job.Spec.CompletionMode == batchv1.IndexedCompletion {
		for i := range job.Spec.Template.Spec.Containers {
			job.Spec.Template.Spec.Containers[i].Env = append(
				job.Spec.Template.Spec.Containers[i].Env, corev1.EnvVar{
					Name:  "KRUN_WAIT_MIN",
					Value: strconv.Itoa(int(*job.Spec.Parallelism)),
				},
			)
		}
	}

	return nil
}
