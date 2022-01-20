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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// BookstoreSpec defines the desired state of BookstoreCRD
type BookstoreSpec struct {
	Name          string `json:"name"`
	ReplicaCount  *int32 `json:"replicaCount"`
	ContainerPort int32  `json:"containerPort"`
}

// BookstoreStatus defines the observed state of BookstoreCRD
type BookstoreStatus struct {
	AvailableReplicas int32 `json:"availableReplicas"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Bookstore is the Schema for the bookstores API
type Bookstore struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec BookstoreSpec `json:"spec"`
	// +optional
	Status BookstoreStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// BookstoreList contains a list of Bookstore
type BookstoreList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Bookstore `json:"items"`
}
