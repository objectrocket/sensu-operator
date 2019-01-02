package v1beta1

import (
	sensu_api_core_v2 "github.com/sensu/sensu-go/api/core/v2"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CheckConfig is the k8s object associated with a sensu check
type CheckConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              CheckConfigSpec `json:"spec"`
}

// CheckConfigSpec is the k8s specification of a sensu check
type CheckConfigSpec struct {
	CheckConfig sensu_api_core_v2.CheckConfig
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CheckConfigList is a list of CheckConfigs.
type CheckConfigList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata
	// More info: http://releases.k8s.io/HEAD/docs/devel/api-conventions.md#metadata
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CheckConfig `json:"items"`
}
