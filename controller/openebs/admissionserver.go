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

package openebs

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"mayadata.io/openebs-upgrade/types"
	"mayadata.io/openebs-upgrade/unstruct"
)

const (
	// DefaultAdmissionServerReplicaCount is the default value of replica for
	// Admission server.
	DefaultAdmissionServerReplicaCount int32 = 1
)

// Set the admission server default values if not already set
func (p *Planner) setAdmissionServerDefaultsIfNotSet() error {
	// Initialize admissionserver field if not set
	if p.ObservedOpenEBS.Spec.AdmissionServer == nil {
		p.ObservedOpenEBS.Spec.AdmissionServer = &types.AdmissionServer{}
	}
	// Enable admission server if the field is not set i.e. set the
	// value to true.
	// TODO: Validate the values that can be provided for this
	// field.
	if p.ObservedOpenEBS.Spec.AdmissionServer.Enabled == nil {
		p.ObservedOpenEBS.Spec.AdmissionServer.Enabled = new(bool)
		*p.ObservedOpenEBS.Spec.AdmissionServer.Enabled = true
	}
	// set the name which will be used to deploy
	if len(p.ObservedOpenEBS.Spec.AdmissionServer.Name) == 0 {
		p.ObservedOpenEBS.Spec.AdmissionServer.Name = types.AdmissionServerNameKey
	}
	if p.ObservedOpenEBS.Spec.AdmissionServer.ImageTag == "" {
		p.ObservedOpenEBS.Spec.AdmissionServer.ImageTag = p.ObservedOpenEBS.Spec.Version +
			p.ObservedOpenEBS.Spec.ImageTagSuffix
	}
	p.ObservedOpenEBS.Spec.AdmissionServer.Image = p.ObservedOpenEBS.Spec.ImagePrefix +
		"admission-server:" + p.ObservedOpenEBS.Spec.AdmissionServer.ImageTag

	if p.ObservedOpenEBS.Spec.AdmissionServer.Replicas == nil {
		p.ObservedOpenEBS.Spec.AdmissionServer.Replicas = new(int32)
		*p.ObservedOpenEBS.Spec.AdmissionServer.Replicas = DefaultAdmissionServerReplicaCount
	}
	return nil
}

// updateAdmissionServer updates the openebs-admission-server manifest as per the
// reconcile.ObservedOpenEBS values.
func (p *Planner) updateAdmissionServer(deploy *unstructured.Unstructured) error {
	deploy.SetName(p.ObservedOpenEBS.Spec.AdmissionServer.Name)
	// desiredLabels is used to form the desired labels of a particular OpenEBS component.
	desiredLabels := deploy.GetLabels()
	if desiredLabels == nil {
		desiredLabels = make(map[string]string, 0)
	}
	// Component specific labels for openebs-snapshot-operator deploy
	// 1. openebs-upgrade.dao.mayadata.io/component-name: openebs-admission-server
	desiredLabels[types.OpenEBSComponentNameLabelKey] = types.AdmissionServerNameKey
	// set the desired labels
	deploy.SetLabels(desiredLabels)
	containers, err := unstruct.GetNestedSliceOrError(deploy, "spec", "template", "spec", "containers")
	if err != nil {
		return err
	}
	// update the containers
	updateContainer := func(obj *unstructured.Unstructured) error {
		envs, _, err := unstruct.GetSlice(obj, "spec", "env")
		if err != nil {
			return err
		}
		containerName, _, err := unstructured.NestedString(obj.Object, "spec", "name")
		if err != nil {
			return err
		}
		if containerName == "admission-webhook" {
			// update the container name if not same.
			if len(p.ObservedOpenEBS.Spec.AdmissionServer.ContainerName) != 0 {
				err = unstructured.SetNestedField(obj.Object, p.ObservedOpenEBS.Spec.AdmissionServer.ContainerName, "spec", "name")
				if err != nil {
					return err
				}
			}
			// Set the image of the container.
			err = unstructured.SetNestedField(obj.Object, p.ObservedOpenEBS.Spec.AdmissionServer.Image,
				"spec", "image")
			if err != nil {
				return err
			}
			// ignore updating the Envs which could cause immutability error
			envs, err = p.ignoreUpdatingImmutableEnvs(p.ObservedOpenEBS.Spec.AdmissionServer.ENV, envs)
			if err != nil {
				return err
			}
		}
		err = unstructured.SetNestedSlice(obj.Object, envs, "spec", "env")
		if err != nil {
			return err
		}
		return nil
	}
	// Update the containers.
	err = unstruct.SliceIterator(containers).ForEachUpdate(updateContainer)
	if err != nil {
		return err
	}
	// Set back the value of the containers.
	err = unstructured.SetNestedSlice(deploy.Object,
		containers, "spec", "template", "spec", "containers")
	if err != nil {
		return err
	}

	return nil
}

func (p *Planner) fillAdmissionServerExistingValues(observedComponentDetails ObservedComponentDesiredDetails) error {
	var (
		containerName string
		err           error
	)
	p.ObservedOpenEBS.Spec.AdmissionServer.MatchLabels = observedComponentDetails.MatchLabels
	p.ObservedOpenEBS.Spec.AdmissionServer.PodTemplateLabels = observedComponentDetails.PodTemplateLabels
	if len(p.ObservedOpenEBS.Spec.AdmissionServer.ContainerName) > 0 {
		containerName = p.ObservedOpenEBS.Spec.AdmissionServer.ContainerName
	} else {
		containerName = types.AdmissionServerContainerKey
	}
	p.ObservedOpenEBS.Spec.AdmissionServer.ENV, err = fetchExistingContainerEnvs(
		observedComponentDetails.Containers, containerName)
	if err != nil {
		return err
	}

	return nil
}
