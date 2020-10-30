/*
Copyright 2020 GramLabs, Inc.

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
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func init() {
	localSchemeBuilder.Register(RegisterDefaults)
}

// Register the defaulting function for the application root object.
func RegisterDefaults(s *runtime.Scheme) error {
	s.AddTypeDefaultingFunc(&Application{}, func(obj interface{}) { obj.(*Application).Default() })
	return nil
}

var _ admission.Defaulter = &Application{}

func (in *Application) Default() {
	// Give scenarios a default name
	for i := range in.Scenarios {
		if in.Scenarios[i].Name == "" {
			in.Scenarios[i].Name = "default"
		}
	}

	// Give objectives a default name based on their type
	for i := range in.Objectives {
		if in.Objectives[i].Name == "" {
			switch {
			case in.Objectives[i].Cost != nil:
				in.Objectives[i].Name = "cost"
			case in.Objectives[i].Latency != nil:
				in.Objectives[i].Name = "latency"
			}
		}
	}
}
