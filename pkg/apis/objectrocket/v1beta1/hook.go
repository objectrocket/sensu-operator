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
)

// HookList is a list of Hooks.
type HookList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata
	// More info: http://releases.k8s.io/HEAD/docs/devel/api-conventions.md#metadata
	metav1.ListMeta `json:"metadata,omitempty"`
	Hooks           []string `json:"hooks"`
}

// A Hook is a hook specification and optionally the results of the hook's
// execution.
// type Hook struct {
// 	// Config is the specification of a hook
// 	HookConfig `json:""`
// 	// Duration of execution
// 	Duration float64 `json:"duration,omitempty"`
// 	// Executed describes the time in which the hook request was executed
// 	Executed int64 `json:"executed"`
// 	// Issued describes the time in which the hook request was issued
// 	Issued int64 `json:"issued"`
// 	// Output from the execution of Command
// 	Output string `json:"output,omitempty"`
// 	// Status is the exit status code produced by the hook
// 	Status int32 `json:"status"`
// }

// // HookConfig is the configuration for a sensu hook
// type HookConfig struct {
// 	// Command is the command to be executed
// 	Command string `json:"command,omitempty"`
// 	// Timeout is the timeout, in seconds, at which the hook has to run
// 	Timeout uint32 `json:"timeout"`
// }
