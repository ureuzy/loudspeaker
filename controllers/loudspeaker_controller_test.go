package controllers

import (
	"context"
	"errors"
	"k8s.io/utils/pointer"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	loudspeakerv1alpha1 "github.com/ureuzy/loudspeaker/api/v1alpha1"
)

var _ = Describe("Loudspeaker controller", func() {
	ctx := context.Background()
	var stopFunc func()

	BeforeEach(func() {
		err := k8sClient.DeleteAllOf(ctx, &loudspeakerv1alpha1.Loudspeaker{}, client.InNamespace("test"))
		Expect(err).NotTo(HaveOccurred())
		err = k8sClient.DeleteAllOf(ctx, &corev1.ConfigMap{}, client.InNamespace("test"))
		Expect(err).NotTo(HaveOccurred())
		err = k8sClient.DeleteAllOf(ctx, &appsv1.Deployment{}, client.InNamespace("test"))
		Expect(err).NotTo(HaveOccurred())
		svcs := &corev1.ServiceList{}
		err = k8sClient.List(ctx, svcs, client.InNamespace("test"))
		Expect(err).NotTo(HaveOccurred())
		for _, svc := range svcs.Items {
			err := k8sClient.Delete(ctx, &svc)
			Expect(err).NotTo(HaveOccurred())
		}
		time.Sleep(100 * time.Millisecond)

		mgr, err := ctrl.NewManager(cfg, ctrl.Options{
			Scheme: scheme,
		})
		Expect(err).ToNot(HaveOccurred())

		reconciler := LoudspeakerReconciler{
			Client:   k8sClient,
			Scheme:   scheme,
			Recorder: mgr.GetEventRecorderFor("loudspeaker-controller"),
		}
		err = reconciler.SetupWithManager(mgr)
		Expect(err).NotTo(HaveOccurred())

		ctx, cancel := context.WithCancel(ctx)
		stopFunc = cancel
		go func() {
			err := mgr.Start(ctx)
			if err != nil {
				panic(err)
			}
		}()
		time.Sleep(100 * time.Millisecond)
	})

	AfterEach(func() {
		stopFunc()
		time.Sleep(100 * time.Millisecond)
	})

	Context("Create Configmap", func() {
		// TODO
		//It("should number of configmaps the same number of listeners", func() {
		//	loudspeaker := newLoudSpeaker()
		//	err := k8sClient.Create(ctx, loudspeaker)
		//	Expect(err).NotTo(HaveOccurred())
		//
		//	cmList := corev1.ConfigMapList{}
		//	opt := &client.ListOptions{
		//		LabelSelector: labels.SelectorFromSet(labelSet(*loudspeaker)),
		//		Namespace:     loudspeaker.Namespace,
		//	}
		//	err = k8sClient.List(ctx, &cmList, opt)
		//	Expect(err).NotTo(HaveOccurred())
		//	Expect(len(cmList.Items)).Should(Equal(len(loudspeaker.Spec.Listeners)))
		//})

		It("should create ConfigMap", func() {
			loudspeaker := newLoudSpeaker()
			err := k8sClient.Create(ctx, loudspeaker)
			Expect(err).NotTo(HaveOccurred())

			cm := corev1.ConfigMap{}
			Eventually(func() error {
				return k8sClient.Get(ctx, client.ObjectKey{Namespace: "test", Name: "sample-foo"}, &cm)
			}).Should(Succeed())
			Expect(cm.Data).Should(HaveKey("observes"))

			Eventually(func() error {
				return k8sClient.Get(ctx, client.ObjectKey{Namespace: "test", Name: "sample-bar"}, &cm)
			}).Should(Succeed())
			Expect(cm.Data).Should(HaveKey("observes"))
		})
	})

	Context("Create Deployment", func() {
		// TODO
		//It("should number of deployments the same number of listeners", func() {
		//	loudspeaker := newLoudSpeaker()
		//	err := k8sClient.Create(ctx, loudspeaker)
		//	Expect(err).NotTo(HaveOccurred())
		//
		//	depList := appsv1.DeploymentList{}
		//	opt := &client.ListOptions{
		//		LabelSelector: labels.SelectorFromSet(labelSet(*loudspeaker)),
		//		Namespace:     loudspeaker.Namespace,
		//	}
		//	err = k8sClient.List(ctx, &depList, opt)
		//	Expect(err).NotTo(HaveOccurred())
		//	Expect(len(depList.Items)).Should(Equal(len(loudspeaker.Spec.Listeners)))
		//})

		It("should create Deployment", func() {
			loudspeaker := newLoudSpeaker()
			err := k8sClient.Create(ctx, loudspeaker)
			Expect(err).NotTo(HaveOccurred())

			dep := appsv1.Deployment{}
			Eventually(func() error {
				return k8sClient.Get(ctx, client.ObjectKey{Namespace: "test", Name: "sample-foo"}, &dep)
			}).Should(Succeed())
			Expect(dep.Spec.Template.Spec.Containers[0].Env[0].Name).Should(Equal("CONFIGMAP"))
			Expect(dep.Spec.Template.Spec.Containers[0].Env[0].Value).Should(Equal("sample-foo"))
			Expect(dep.Spec.Replicas).Should(Equal(pointer.Int32Ptr(1)))
			Expect(dep.Spec.Template.Spec.Containers[0].Image).Should(Equal("nginx:latest"))

			Eventually(func() error {
				return k8sClient.Get(ctx, client.ObjectKey{Namespace: "test", Name: "sample-bar"}, &dep)
			}).Should(Succeed())
			Expect(dep.Spec.Template.Spec.Containers[0].Env[0].Name).Should(Equal("CONFIGMAP"))
			Expect(dep.Spec.Template.Spec.Containers[0].Env[0].Value).Should(Equal("sample-bar"))
			Expect(dep.Spec.Replicas).Should(Equal(pointer.Int32Ptr(1)))
			Expect(dep.Spec.Template.Spec.Containers[0].Image).Should(Equal("nginx:latest"))
		})
	})

	It("should update status", func() {
		loudspeaker := newLoudSpeaker()
		err := k8sClient.Create(ctx, loudspeaker)
		Expect(err).NotTo(HaveOccurred())

		updated := loudspeakerv1alpha1.Loudspeaker{}
		Eventually(func() error {
			err := k8sClient.Get(ctx, client.ObjectKey{Namespace: "test", Name: "sample"}, &updated)
			if err != nil {
				return err
			}
			if updated.Status.Status == "" {
				return errors.New("status should be updated")
			}
			return nil
		}).Should(Succeed())
	})
})

func newLoudSpeaker() *loudspeakerv1alpha1.Loudspeaker {
	return &loudspeakerv1alpha1.Loudspeaker{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "sample",
			Namespace: "test",
		},
		Spec: loudspeakerv1alpha1.LoudspeakerSpec{
			Image:              "nginx:latest",
			ServiceAccountName: "sample-service-account",
			Listeners: []loudspeakerv1alpha1.Listener{
				{
					Name:        "foo",
					Type:        "sentry",
					Credentials: "sample-secrets",
					Observes: []loudspeakerv1alpha1.Observe{
						{
							Namespace:         "default",
							IgnoreReasons:     []string{"BackoffLimitExceeded"},
							IgnoreObjectNames: []string{"sample-foo"},
							IgnoreObjectKinds: []string{"Deployment"},
							IgnoreEventTypes:  []string{"Normal"},
						},
					},
				},
				{
					Name:        "bar",
					Type:        "sentry",
					Credentials: "sample-secrets",
					Observes: []loudspeakerv1alpha1.Observe{
						{
							Namespace:         "default",
							IgnoreReasons:     []string{"BackoffLimitExceeded"},
							IgnoreObjectNames: []string{"sample-foo"},
							IgnoreObjectKinds: []string{"Deployment"},
							IgnoreEventTypes:  []string{"Normal"},
						},
					},
				},
			},
		},
	}
}
