/*
Copyright 2020 The CRDS Authors.

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

package crd

import (
	"encoding/json"
	"errors"
	"fmt"

	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	servervalidation "k8s.io/apiextensions-apiserver/pkg/apiserver/validation"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/validation"
	"sigs.k8s.io/yaml"
)

const (
	getStoredGVKErr          = "unable to determind stored gvk"
	createSchemaValidatorErr = "could not create schema validator"
	yamlToJSONErr            = "could not convert yaml to json"
	getTypeMetaErr           = "could not get type metadata for crd instance"
	wrongGVKErr              = "crd instance was not of correct group version kind"
	instanceConversionErr    = "could not convert crd instance json to instance"
)

// CRDer generates instances of a CustomResourceDefinition.
type CRDer struct {
	CRD *apiextensions.CustomResourceDefinition
	GVK *schema.GroupVersionKind
}

// NewCRDer returns a new CRDer type.
func NewCRDer(data []byte, m ...Modifier) (*CRDer, error) {
	internal := &apiextensions.CustomResourceDefinition{}
	if errV1Beta1 := convertV1Beta1ToInternal(data, internal, m...); errV1Beta1 != nil {
		if errV1 := convertV1ToInternal(data, internal, m...); errV1 != nil {
			return nil, fmt.Errorf("conversion unsuccessful: %s, %s", errV1Beta1, errV1)
		}
	}

	gvk := GetStoredGVK(internal)
	if gvk == nil {
		return nil, errors.New(getStoredGVKErr)
	}

	return &CRDer{CRD: internal, GVK: gvk}, nil
}

// Validate returns true if CRD instance is valid.
func (c *CRDer) Validate(data []byte) error {
	sv := getStoredSchema(c.CRD.Spec)

	s, _, err := servervalidation.NewSchemaValidator(sv)
	if err != nil {
		return errors.New(createSchemaValidatorErr)
	}

	j, err := yaml.YAMLToJSONStrict(data)
	if err != nil {
		return errors.New(yamlToJSONErr)
	}

	meta := &metav1.TypeMeta{}
	if err := json.Unmarshal(j, meta); err != nil {
		return errors.New(getTypeMetaErr)
	}

	if !isStoredGVK(meta, c.GVK) {
		return errors.New(wrongGVKErr)
	}

	var instance interface{}
	if err := json.Unmarshal(j, &instance); err != nil {
		return errors.New(instanceConversionErr)
	}

	res := servervalidation.ValidateCustomResource(nil, instance, s)
	if len(res) > 0 {
		return errors.New(res.ToAggregate().Error())
	}
	return nil
}

func convertV1ToInternal(data []byte, internal *apiextensions.CustomResourceDefinition, mods ...Modifier) error {
	crd := &v1.CustomResourceDefinition{}
	if err := yaml.Unmarshal(data, crd); err != nil {
		return err
	}
	v1.SetDefaults_CustomResourceDefinition(crd)
	if err := v1.Convert_v1_CustomResourceDefinition_To_apiextensions_CustomResourceDefinition(crd, internal, nil); err != nil {
		return err
	}
	for _, m := range mods {
		m(internal)
	}
	errList := validation.ValidateCustomResourceDefinition(internal, v1.SchemeGroupVersion)
	if len(errList) > 0 {
		return errors.New(errList.ToAggregate().Error())
	}

	return nil
}

func convertV1Beta1ToInternal(data []byte, internal *apiextensions.CustomResourceDefinition, mods ...Modifier) error {
	crd := &v1beta1.CustomResourceDefinition{}
	if err := yaml.Unmarshal(data, crd); err != nil {
		return err
	}
	v1beta1.SetObjectDefaults_CustomResourceDefinition(crd)
	if err := v1beta1.Convert_v1beta1_CustomResourceDefinition_To_apiextensions_CustomResourceDefinition(crd, internal, nil); err != nil {
		return err
	}
	for _, m := range mods {
		m(internal)
	}
	errList := validation.ValidateCustomResourceDefinition(internal, v1beta1.SchemeGroupVersion)
	if len(errList) > 0 {
		return errors.New(errList.ToAggregate().Error())
	}

	return nil
}

func getStoredSchema(spec apiextensions.CustomResourceDefinitionSpec) *apiextensions.CustomResourceValidation {
	if spec.Validation != nil {
		return spec.Validation
	}
	for _, v := range spec.Versions {
		if v.Storage {
			return v.Schema
		}
	}
	return nil
}

func GetStoredGVK(crd *apiextensions.CustomResourceDefinition) *schema.GroupVersionKind {
	for _, v := range crd.Spec.Versions {
		if v.Storage {
			return &schema.GroupVersionKind{
				Group:   crd.Spec.Group,
				Version: v.Name,
				Kind:    crd.Spec.Names.Kind,
			}
		}
	}

	return nil
}

func isStoredGVK(meta *metav1.TypeMeta, gvk *schema.GroupVersionKind) bool {
	if meta.GroupVersionKind() == *gvk {
		return true
	}

	return false
}

// A Modifier specifies how to modify a CRD prior to conversion to internal
// representation
type Modifier func(crd *apiextensions.CustomResourceDefinition)

// StripLabels removes labels from a CRD's metadata
func StripLabels() Modifier {
	return func(crd *apiextensions.CustomResourceDefinition) {
		crd.SetLabels(map[string]string{})
	}
}

// StripAnnotations removes annotations from a CRD's metadata
func StripAnnotations() Modifier {
	return func(crd *apiextensions.CustomResourceDefinition) {
		crd.SetAnnotations(map[string]string{})
	}
}

// StripConversion removes conversion from a CRD's spec
func StripConversion() Modifier {
	return func(crd *apiextensions.CustomResourceDefinition) {
		crd.Spec.Conversion = nil
	}
}

// PrettyGVK returns a group, version, kind representation in order of
// specificity.
func PrettyGVK(gvk *schema.GroupVersionKind) string {
	return fmt.Sprintf("%s/%s/%s", gvk.Group, gvk.Kind, gvk.Version)
}
