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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/json"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ListenerType +kubebuilder:validation:Enum=sentry
type ListenerType string

const (
	Sentry = ListenerType("sentry")
)

type Watch struct {
	//+kubebuilder:validation:Required
	Namespace string `json:"namespace"`

	Ignore []string `json:"ignore,omitempty"`
}

type Listener struct {
	//+kubebuilder:validation:Required
	Type ListenerType `json:"type,omitempty"`
	//+kubebuilder:validation:Required
	Credentials string `json:"credentials"`
	//+kubebuilder:validation:Required
	Watch []Watch `json:"watch,omitempty"`
}

type Listeners []Listener

// LoudspeakerSpec defines the desired state of Loudspeaker
type LoudspeakerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of Loudspeaker. Edit loudspeaker_types.go to remove/update

	//+kubebuilder:validation:Required
	Listeners Listeners `json:"listeners,omitempty"`
	//+kubebuilder:validation:Required
	Image string `json:"image"`
}

// LoudspeakerStatus +kubebuilder:validation:Enum=NotReady;Available;Healthy
type LoudspeakerStatus string

const (
	LoudspeakerNotReady  = LoudspeakerStatus("NotReady")
	LoudspeakerAvailable = LoudspeakerStatus("Available")
	LoudspeakerHealthy   = LoudspeakerStatus("Healthy")
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:shortName=lo
//+kubebuilder:printcolumn:name="IMAGE",type="string",JSONPath=".spec.image"

// Loudspeaker is the Schema for the loudspeakers API
type Loudspeaker struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LoudspeakerSpec   `json:"spec,omitempty"`
	Status LoudspeakerStatus `json:"status,omitempty"`
}

func (l *Loudspeaker) ToJsonString() (string, error) {
	b, err := json.Marshal(l.Spec)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

//+kubebuilder:object:root=true

// LoudspeakerList contains a list of Loudspeaker
type LoudspeakerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Loudspeaker `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Loudspeaker{}, &LoudspeakerList{})
}
