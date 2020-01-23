/*
Copyright 2020 The MayaData Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    https://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package k8s

import (
	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Namespace is a wrapper over k8s Namespace
type Namespace struct {
	Object *corev1.Namespace `json:"object"`
}

// createOrUpdate checks if the resource provided is present or not, if
// not present then it creates the resource otherwise updates it.
func (namespace *Namespace) createOrUpdate() error {
	existingNs, err := Clientset.CoreV1().Namespaces().Get(namespace.Object.Name, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			_, err = Clientset.CoreV1().Namespaces().Create(namespace.Object)
			if err != nil {
				return errors.Errorf("Error creating namespace: %s: %+v", namespace.Object.Name, err)
			}
		} else {
			return errors.Errorf("Error getting namespace: %s: %+v", namespace.Object.Name, err)
		}
	}
	// Set the resource version of the object to be updated
	namespace.Object.SetResourceVersion(existingNs.GetResourceVersion())
	_, err = Clientset.CoreV1().Namespaces().Update(namespace.Object)
	if err != nil {
		return errors.Errorf("Error updating namespace: %s: %+v", namespace.Object.Name, err)
	}
	return nil
}

// DeployNamespace creates/updates a given namespace based on
// the given YAML.
func DeployNamespace(YAML string) error {
	ns := &corev1.Namespace{}
	err := yaml.Unmarshal([]byte(YAML), ns)
	if err != nil {
		return errors.Errorf(
			"Error unmarshalling namespace YAML: %+v", err)
	}
	namespace := &Namespace{
		Object: ns,
	}
	err = namespace.createOrUpdate()
	if err != nil {
		return err
	}
	return nil
}
