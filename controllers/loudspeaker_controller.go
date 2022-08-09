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
	"fmt"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	appsv1apply "k8s.io/client-go/applyconfigurations/apps/v1"
	corev1apply "k8s.io/client-go/applyconfigurations/core/v1"
	metav1apply "k8s.io/client-go/applyconfigurations/meta/v1"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/yaml"

	loudspeakerv1alpha1 "github.com/ureuzy/loudspeaker/api/v1alpha1"
	"github.com/ureuzy/loudspeaker/pkg/constants"
)

// LoudspeakerReconciler reconciles a Loudspeaker object
type LoudspeakerReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=loudspeaker.ureuzy.github.io,resources=loudspeakers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=loudspeaker.ureuzy.github.io,resources=loudspeakers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=loudspeaker.ureuzy.github.io,resources=loudspeakers/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups=core,resources=events,verbs=create;update;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Loudspeaker object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *LoudspeakerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var loudspeaker loudspeakerv1alpha1.Loudspeaker
	err := r.Get(ctx, req.NamespacedName, &loudspeaker)
	if errors.IsNotFound(err) {
		//r.removeMetrics(loudspeaker)
		return ctrl.Result{}, nil
	}
	if err != nil {
		logger.Error(err, "unable to get Loudspeaker")
		return ctrl.Result{}, err
	}

	if !loudspeaker.ObjectMeta.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, nil
	}

	for _, listener := range loudspeaker.Spec.Listeners {
		if err = r.reconcileConfigMap(ctx, loudspeaker, listener); err != nil {
			return ctrl.Result{}, err
		}

		if err = r.reconcileDeployment(ctx, loudspeaker, listener); err != nil {
			return ctrl.Result{}, err
		}
	}

	if err = r.reconcileGarbageCollection(ctx, loudspeaker); err != nil {
		return ctrl.Result{}, err
	}

	return r.updateStatus(ctx, loudspeaker)
}

func (r *LoudspeakerReconciler) reconcileConfigMap(ctx context.Context, loudspeaker loudspeakerv1alpha1.Loudspeaker, listener loudspeakerv1alpha1.Listener) error {
	logger := log.FromContext(ctx)

	cm := &corev1.ConfigMap{}
	cm.SetNamespace(loudspeaker.Namespace)
	cm.SetName(generateName(loudspeaker, listener))
	cm.SetLabels(labelSet(loudspeaker))

	op, err := ctrl.CreateOrUpdate(ctx, r.Client, cm, func() error {
		if cm.Data == nil {
			cm.Data = make(map[string]string)
		}

		obj, err := yaml.Marshal(listener.Observes)
		if err != nil {
			return err
		}

		cm.Data["observes"] = string(obj)
		return ctrl.SetControllerReference(&loudspeaker, cm, r.Scheme)
	})

	if err != nil {
		logger.Error(err, "unable to create or update ConfigMap")
		return err
	}

	if op != controllerutil.OperationResultNone {
		logger.Info("reconcile ConfigMap successfully", "op", op)
	}
	return nil
}

func (r *LoudspeakerReconciler) reconcileDeployment(ctx context.Context, loudspeaker loudspeakerv1alpha1.Loudspeaker, listener loudspeakerv1alpha1.Listener) error {
	logger := log.FromContext(ctx)

	owner, err := ownerRef(loudspeaker, r.Scheme)
	if err != nil {
		return err
	}

	name := generateName(loudspeaker, listener)

	volume := corev1apply.VolumeApplyConfiguration{
		Name: &listener.Credentials,
		VolumeSourceApplyConfiguration: corev1apply.VolumeSourceApplyConfiguration{
			Secret: &corev1apply.SecretVolumeSourceApplyConfiguration{
				SecretName: &listener.Credentials,
			},
		},
	}
	volumeMount := corev1apply.VolumeMountApplyConfiguration{
		Name:      &listener.Credentials,
		ReadOnly:  pointer.Bool(true),
		MountPath: pointer.String(fmt.Sprintf(constants.CredentialsMountPath)),
	}
	volumes := []*corev1apply.VolumeApplyConfiguration{&volume}
	volumeMounts := []*corev1apply.VolumeMountApplyConfiguration{&volumeMount}

	dep := appsv1apply.Deployment(name, loudspeaker.Namespace).
		WithLabels(labelSet(loudspeaker)).
		WithOwnerReferences(owner).
		WithSpec(appsv1apply.DeploymentSpec().
			WithReplicas(constants.DefaultReplicas).
			WithSelector(metav1apply.LabelSelector().WithMatchLabels(labelSet(loudspeaker))).
			WithTemplate(corev1apply.PodTemplateSpec().
				WithLabels(labelSet(loudspeaker)).
				WithSpec(corev1apply.PodSpec().
					WithServiceAccountName(loudspeaker.Spec.ServiceAccountName).
					WithContainers(corev1apply.Container().
						WithName(constants.ContainerName).
						WithImage(loudspeaker.Spec.Image).
						WithImagePullPolicy(corev1.PullIfNotPresent).
						WithVolumeMounts(volumeMounts...).
						WithEnv(
							&corev1apply.EnvVarApplyConfiguration{
								Name:  pointer.String("CONFIGMAP"),
								Value: pointer.String(name),
							},
							&corev1apply.EnvVarApplyConfiguration{
								Name:  pointer.String("TYPE"),
								Value: pointer.String(string(listener.Type)),
							},
							&corev1apply.EnvVarApplyConfiguration{
								Name:  pointer.String("NAMESPACE"),
								Value: pointer.String(loudspeaker.Namespace),
							}),
					).
					WithVolumes(volumes...),
				),
			),
		)

	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(dep)
	if err != nil {
		logger.Error(err, "unstructured")
		return err
	}
	patch := &unstructured.Unstructured{
		Object: obj,
	}

	var current appsv1.Deployment
	err = r.Get(ctx, client.ObjectKey{Namespace: loudspeaker.Namespace, Name: name}, &current)
	if err != nil && !errors.IsNotFound(err) {
		logger.Error(err, "get deployment")
		return err
	}

	currApplyConfig, err := appsv1apply.ExtractDeployment(&current, "loudspeaker-controller")
	if err != nil {
		logger.Error(err, "extractDeployment")
		return err
	}

	if equality.Semantic.DeepEqual(dep, currApplyConfig) {
		return nil
	}

	err = r.Patch(ctx, patch, client.Apply, &client.PatchOptions{
		FieldManager: "loudspeaker-controller",
		Force:        pointer.Bool(true),
	})

	if err != nil {
		logger.Error(err, "unable to create or update Deployment")
		return err
	}

	logger.Info("reconcile Deployment successfully", "name", loudspeaker.Name)

	return nil
}

func (r *LoudspeakerReconciler) reconcileGarbageCollection(ctx context.Context, loudspeaker loudspeakerv1alpha1.Loudspeaker) error {
	logger := log.FromContext(ctx)

	depList := &appsv1.DeploymentList{}
	cmList := &corev1.ConfigMapList{}

	opt := &client.ListOptions{
		LabelSelector: labels.SelectorFromSet(labelSet(loudspeaker)),
		Namespace:     loudspeaker.Namespace,
	}

	err := r.List(ctx, depList, opt)
	if err != nil {
		return err
	}

	err = r.List(ctx, cmList, opt)
	if err != nil {
		return err
	}

	for _, dep := range depList.Items {
		listenerName := extractListenerName(dep.Name, loudspeaker)
		if loudspeaker.IncludeListener(listenerName) {
			continue
		}
		err = r.Delete(ctx, &dep)
		if err != nil {
			logger.Error(err, fmt.Sprintf("delete deployment: %s", dep.Name))
		}
	}

	for _, cm := range cmList.Items {
		listenerConfigName := extractListenerName(cm.Name, loudspeaker)
		if loudspeaker.IncludeListener(listenerConfigName) {
			continue
		}
		err = r.Delete(ctx, &cm)
		if err != nil {
			logger.Error(err, fmt.Sprintf("delete Configmap: %s", cm.Name))
		}
	}

	return nil
}

func (r *LoudspeakerReconciler) updateStatus(ctx context.Context, loudspeaker loudspeakerv1alpha1.Loudspeaker) (ctrl.Result, error) {

	depList := &appsv1.DeploymentList{}
	opt := &client.ListOptions{
		LabelSelector: labels.SelectorFromSet(labelSet(loudspeaker)),
		Namespace:     loudspeaker.Namespace,
	}
	err := r.List(ctx, depList, opt)
	if err != nil {
		return ctrl.Result{}, err
	}

	availableDeployments := 0
	for _, dep := range depList.Items {
		if dep.Status.AvailableReplicas > 0 {
			availableDeployments += 1
		}
	}

	var status loudspeakerv1alpha1.LoudspeakerStatus
	if availableDeployments == 0 {
		status.Status = loudspeakerv1alpha1.LoudspeakerNotReady
	} else if availableDeployments == len(loudspeaker.Spec.Listeners) {
		status.Status = loudspeakerv1alpha1.LoudspeakerHealthy
	} else {
		status.Status = loudspeakerv1alpha1.LoudspeakerAvailable
	}
	status.AvailableListener = fmt.Sprintf("%d/%d", availableDeployments, len(depList.Items))

	if loudspeaker.Status.Status != status.Status || loudspeaker.Status.AvailableListener != status.AvailableListener {
		loudspeaker.Status = status
		r.setMetrics(loudspeaker)

		r.Recorder.Event(&loudspeaker, corev1.EventTypeNormal, "Updated", fmt.Sprintf("Loudspeaker(%s:%s) updated: %s", loudspeaker.Namespace, loudspeaker.Name, loudspeaker.Status.Status))

		err = r.Status().Update(ctx, &loudspeaker)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	if loudspeaker.Status.Status != loudspeakerv1alpha1.LoudspeakerHealthy {
		return ctrl.Result{Requeue: true}, nil
	}
	return ctrl.Result{}, nil
}

func labelSet(loudspeaker loudspeakerv1alpha1.Loudspeaker) map[string]string {
	return map[string]string{constants.RuntimeLabelsKey: loudspeaker.Name}
}

func generateName(loudspeaker loudspeakerv1alpha1.Loudspeaker, listener loudspeakerv1alpha1.Listener) string {
	return fmt.Sprintf("%s-%s", loudspeaker.Name, listener.Name)
}

func extractListenerName(resourceName string, loudspeaker loudspeakerv1alpha1.Loudspeaker) string {
	return strings.Replace(resourceName, loudspeaker.Name+"-", "", 1)
}

func (r *LoudspeakerReconciler) setMetrics(loudspeaker loudspeakerv1alpha1.Loudspeaker) {
	switch loudspeaker.Status.Status {
	case loudspeakerv1alpha1.LoudspeakerNotReady:
		//metrics.NotReadyVec.WithLabelValues(loudspeaker.Name, loudspeaker.Name).Set(1)
		//metrics.AvailableVec.WithLabelValues(loudspeaker.Name, loudspeaker.Name).Set(0)
		//metrics.HealthyVec.WithLabelValues(loudspeaker.Name, loudspeaker.Name).Set(0)
	case loudspeakerv1alpha1.LoudspeakerAvailable:
		//metrics.NotReadyVec.WithLabelValues(loudspeaker.Name, loudspeaker.Name).Set(0)
		//metrics.AvailableVec.WithLabelValues(loudspeaker.Name, loudspeaker.Name).Set(1)
		//metrics.HealthyVec.WithLabelValues(loudspeaker.Name, loudspeaker.Name).Set(0)
	case loudspeakerv1alpha1.LoudspeakerHealthy:
		//metrics.NotReadyVec.WithLabelValues(loudspeaker.Name, loudspeaker.Name).Set(0)
		//metrics.AvailableVec.WithLabelValues(loudspeaker.Name, loudspeaker.Name).Set(0)
		//metrics.HealthyVec.WithLabelValues(loudspeaker.Name, loudspeaker.Name).Set(1)
	}
}

func ownerRef(loudspeaker loudspeakerv1alpha1.Loudspeaker, scheme *runtime.Scheme) (*metav1apply.OwnerReferenceApplyConfiguration, error) {
	gvk, err := apiutil.GVKForObject(&loudspeaker, scheme)
	if err != nil {
		return nil, err
	}
	ref := metav1apply.OwnerReference().
		WithAPIVersion(gvk.GroupVersion().String()).
		WithKind(gvk.Kind).
		WithName(loudspeaker.Name).
		WithUID(loudspeaker.GetUID()).
		WithBlockOwnerDeletion(true).
		WithController(true)
	return ref, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *LoudspeakerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&loudspeakerv1alpha1.Loudspeaker{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}
