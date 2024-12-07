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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// SSHKeyPairSpec defines the desired state of SSHKeyPair.
type SSHKeyPairSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	PublicKey  string `json:"publicKey,omitempty"`
	PrivateKey string `json:"privateKey,omitempty"`
	Type       string `json:"type,omitempty"`
	// Name describes the stored file name of the SSH keys;
	// for example, if the name is `my-key`, the public key
	// will be stored in `my-key.pub` and the private key
	// will be stored in `my-key`.
	Name string `json:"name,omitempty"`
}

// SSHKeyPairStatus defines the observed state of SSHKeyPair.
type SSHKeyPairStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// SSHKeyPair is the Schema for the sshkeypairs API.
type SSHKeyPair struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SSHKeyPairSpec   `json:"spec,omitempty"`
	Status SSHKeyPairStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SSHKeyPairList contains a list of SSHKeyPair.
type SSHKeyPairList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SSHKeyPair `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SSHKeyPair{}, &SSHKeyPairList{})
}
