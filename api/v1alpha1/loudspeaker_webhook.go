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

package v1alpha1

import (
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	"github.com/masanetes/loudspeaker/pkg/constants"
)

// log is for logging in this package.
var loudspeakerlog = logf.Log.WithName("loudspeaker-resource")

func (r *Loudspeaker) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-loudspeaker-masanetes-github-io-v1alpha1-loudspeaker,mutating=true,failurePolicy=fail,sideEffects=None,groups=loudspeaker.masanetes.github.io,resources=loudspeakers,verbs=create;update,versions=v1alpha1,name=mloudspeaker.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.Defaulter = &Loudspeaker{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Loudspeaker) Default() {
	loudspeakerlog.Info("default", "name", r.Name)

	if len(r.Spec.Image) == 0 {
		r.Spec.Image = constants.DefaultImage
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-loudspeaker-masanetes-github-io-v1alpha1-loudspeaker,mutating=false,failurePolicy=fail,sideEffects=None,groups=loudspeaker.masanetes.github.io,resources=loudspeakers,verbs=create;update,versions=v1alpha1,name=vloudspeaker.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.Validator = &Loudspeaker{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Loudspeaker) ValidateCreate() error {
	loudspeakerlog.Info("validate create", "name", r.Name)

	return r.validateLoudspeaker()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Loudspeaker) ValidateUpdate(old runtime.Object) error {
	loudspeakerlog.Info("validate update", "name", r.Name)

	return r.validateLoudspeaker()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Loudspeaker) ValidateDelete() error {
	loudspeakerlog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}

func (r *Loudspeaker) validateLoudspeaker() error {
	var errs field.ErrorList

	if r.Spec.Listeners.IsDuplicateListenerName() {
		errs = append(errs, field.Duplicate(field.NewPath("spec", "listeners", "name"), "same name must not be specified."))
	}

	if len(errs) > 0 {
		err := apierrors.NewInvalid(schema.GroupKind{Group: GroupVersion.Group, Kind: "Loudspeaker"}, r.Name, errs)
		loudspeakerlog.Error(err, "validation error", "name", r.Name)
		return err
	}

	return nil
}
