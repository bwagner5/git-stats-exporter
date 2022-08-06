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
)

// RepoSpec defines the desired state of Repo
type RepoSpec struct {
	// Owner is the github owner organization or user
	Owner string `json:"owner"`
	// Name is the git repository name
	Name string `json:"name"`
}

// RepoStatus defines the observed state of Repo
type RepoStatus struct {
	State     *string     `json:"state,omitempty"`
	LastQuery metav1.Time `json:"lastQuery,omitempty"`
}

//+kubebuilder:resource:path=repos
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Repo is the Schema for the repos API
type Repo struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RepoSpec   `json:"spec,omitempty"`
	Status RepoStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// RepoList contains a list of Repo
type RepoList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Repo `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Repo{}, &RepoList{})
}
