package v1alpha1

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Loudspeaker Types", func() {

	loudspeaker1 := loudspeaker1()
	loudspeaker2 := loudspeaker2()

	It("Contains ignore reason", func() {
		Expect(loudspeaker1.Spec.Listeners[0].Observes[0].Ignores.Contains("BackoffLimitExceeded")).Should(Equal(true))
	})

	It("No contains ignore reason", func() {
		Expect(loudspeaker1.Spec.Listeners[0].Observes[0].Ignores.Contains("Unhealthy")).Should(Equal(false))
	})

	It("Duplicate listener name", func() {
		Expect(loudspeaker1.Spec.Listeners.DuplicateListenerName()).Should(Equal(true))
	})

	It("No Duplicate listener name", func() {
		Expect(loudspeaker2.Spec.Listeners.DuplicateListenerName()).Should(Equal(false))
	})

	It("Include listener name", func() {
		Expect(loudspeaker2.IncludeListener("foo")).Should(Equal(true))
		Expect(loudspeaker2.IncludeListener("bar")).Should(Equal(true))
	})

	It("No Include listener name", func() {
		Expect(loudspeaker2.IncludeListener("baz")).Should(Equal(false))
	})
})

func loudspeaker1() *Loudspeaker {
	return &Loudspeaker{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "sample",
			Namespace: "test",
		},
		Spec: LoudspeakerSpec{
			Image:              "nginx:latest",
			ServiceAccountName: "sample-service-account",
			Listeners: []Listener{
				{
					Name:        "foo",
					Type:        "sentry",
					Credentials: "sample-secrets",
					Observes: []Observe{
						{
							Namespace: "default",
							Ignores:   []string{"BackoffLimitExceeded"},
						},
					},
				},
				{
					Name:        "foo",
					Type:        "sentry",
					Credentials: "sample-secrets",
					Observes: []Observe{
						{
							Namespace: "default",
							Ignores:   []string{"BackoffLimitExceeded"},
						},
					},
				},
			},
		},
	}
}

func loudspeaker2() *Loudspeaker {
	return &Loudspeaker{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "sample",
			Namespace: "test",
		},
		Spec: LoudspeakerSpec{
			Image:              "nginx:latest",
			ServiceAccountName: "sample-service-account",
			Listeners: []Listener{
				{
					Name:        "foo",
					Type:        "sentry",
					Credentials: "sample-secrets",
					Observes: []Observe{
						{
							Namespace: "default",
							Ignores:   []string{"BackoffLimitExceeded"},
						},
					},
				},
				{
					Name:        "bar",
					Type:        "sentry",
					Credentials: "sample-secrets",
					Observes: []Observe{
						{
							Namespace: "default",
							Ignores:   []string{"BackoffLimitExceeded"},
						},
					},
				},
			},
		},
	}
}
