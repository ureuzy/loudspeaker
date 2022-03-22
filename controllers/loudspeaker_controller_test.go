package controllers

import (
	"context"
	"errors"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	loudspeakerv1 "github.com/masanetes/loudspeaker/api/v1"
)

var _ = Describe("Loudspeaker controller", func() {
	ctx := context.Background()
	var stopFunc func()

	BeforeEach(func() {
		err := k8sClient.DeleteAllOf(ctx, &loudspeakerv1.Loudspeaker{}, client.InNamespace("test"))
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

	It("should create ConfigMap", func() {
		loudspeaker := newLoudSpeaker()
		err := k8sClient.Create(ctx, loudspeaker)
		Expect(err).NotTo(HaveOccurred())

		cm := corev1.ConfigMap{}
		Eventually(func() error {
			return k8sClient.Get(ctx, client.ObjectKey{Namespace: "test", Name: "sample-config"}, &cm)
		}).Should(Succeed())

		Expect(cm.Data).Should(HaveKey("listeners"))
	})

	It("should create Deployment", func() {
		loudspeaker := newLoudSpeaker()
		err := k8sClient.Create(ctx, loudspeaker)
		Expect(err).NotTo(HaveOccurred())

		dep := appsv1.Deployment{}
		Eventually(func() error {
			return k8sClient.Get(ctx, client.ObjectKey{Namespace: "test", Name: "sample-forwarder"}, &dep)
		}).Should(Succeed())
		Expect(dep.Spec.Replicas).Should(Equal(pointer.Int32Ptr(1)))
		Expect(dep.Spec.Template.Spec.Containers[0].Image).Should(Equal("nginx:latest"))
	})

	It("should update status", func() {
		loudspeaker := newLoudSpeaker()
		err := k8sClient.Create(ctx, loudspeaker)
		Expect(err).NotTo(HaveOccurred())

		updated := loudspeakerv1.Loudspeaker{}
		Eventually(func() error {
			err := k8sClient.Get(ctx, client.ObjectKey{Namespace: "test", Name: "sample"}, &updated)
			if err != nil {
				return err
			}
			if updated.Status == "" {
				return errors.New("status should be updated")
			}
			return nil
		}).Should(Succeed())
	})
})

func newLoudSpeaker() *loudspeakerv1.Loudspeaker {
	return &loudspeakerv1.Loudspeaker{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "sample",
			Namespace: "test",
		},
		Spec: loudspeakerv1.LoudspeakerSpec{
			Listeners: []loudspeakerv1.Listener{
				{
					Type:        "sentry",
					Credentials: "sample-secrets",
					Watch: []loudspeakerv1.Watch{
						{
							Namespace: "default",
							Ignore:    []string{"BackoffLimitExceeded"},
						},
					},
				},
			},
			Image: "nginx:latest",
		},
	}
}
