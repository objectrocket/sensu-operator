// Copyright 2017 The etcd-operator Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	SensuClusterResourceKind   = "SensuCluster"
	SensuClusterResourcePlural = "sensuclusters"
	groupName                  = "objectrocket.com"

	SensuAssetResourceKind   = "SensuAsset"
	SensuAssetResourcePlural = "sensuassets"

	SensuCheckConfigResourceKind   = "SensuCheckConfig"
	SensuCheckConfigResourcePlural = "sensucheckconfigs"

	SensuHandlerResourceKind   = "SensuHandler"
	SensuHandlerResourcePlural = "sensuhandlers"

	SensuEventFilterResourceKind   = "SensuEventFilter"
	SensuEventFilterResourcePlural = "sensueventfilters"

	SensuBackupResourceKind   = "SensuBackup"
	SensuBackupResourcePlural = "sensubackups"

	SensuRestoreResourceKind   = "SensuRestore"
	SensuRestoreResourcePlural = "sensurestores"
)

var (
	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)
	AddToScheme   = SchemeBuilder.AddToScheme

	SchemeGroupVersion      = schema.GroupVersion{Group: groupName, Version: "v1beta1"}
	SensuClusterCRDName     = SensuClusterResourcePlural + "." + groupName
	SensuAssetCRDName       = SensuAssetResourcePlural + "." + groupName
	SensuCheckConfigCRDName = SensuCheckConfigResourcePlural + "." + groupName
	SensuHandlerCRDName     = SensuHandlerResourcePlural + "." + groupName
	SensuEventFilterCRDName = SensuEventFilterResourcePlural + "." + groupName
	SensuBackupCRDName      = SensuBackupResourcePlural + "." + groupName
	SensuRestoreCRDName     = SensuRestoreResourcePlural + "." + groupName
)

// Resource gets a SensuCluster GroupResource for a specified resource
func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}

// addKnownTypes adds the set of types defined in this package to the supplied scheme.
func addKnownTypes(s *runtime.Scheme) error {
	s.AddKnownTypes(SchemeGroupVersion,
		&SensuCluster{},
		&SensuClusterList{},
		&SensuBackup{},
		&SensuBackupList{},
		&SensuRestore{},
		&SensuRestoreList{},
		&SensuAsset{},
		&SensuAssetList{},
		&SensuCheckConfig{},
		&SensuCheckConfigList{},
		&SensuHandler{},
		&SensuHandlerList{},
		&SensuEventFilter{},
		&SensuEventFilterList{},
	)
	metav1.AddToGroupVersion(s, SchemeGroupVersion)
	return nil
}
