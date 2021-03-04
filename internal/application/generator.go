/*
Copyright 2021 GramLabs, Inc.

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

package application

import (
	"io"
	"path/filepath"
	"time"

	"github.com/thestormforge/konjure/pkg/konjure"
	redskyappsv1alpha1 "github.com/thestormforge/optimize-controller/api/apps/v1alpha1"
	"github.com/thestormforge/optimize-controller/internal/scan"
	"k8s.io/apimachinery/pkg/util/json"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/kio/filters"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

type Generator struct {
	Name          string
	Resources     konjure.Resources
	Objectives    []string
	DefaultReader io.Reader
}

func (g *Generator) Execute(output kio.Writer) error {
	return kio.Pipeline{
		Inputs: []kio.Reader{g.Resources},
		Filters: []kio.Filter{
			&konjure.Filter{
				Depth:         100,
				DefaultReader: g.DefaultReader,
				KeepStatus:    true,
			},
			&scan.Scanner{
				Selectors:   []scan.Selector{g},
				Transformer: g,
			},
			kio.FilterAll(yaml.Clear("status")),
			// TODO We should have an optional filter that annotates the application with comments
			&filters.FormatFilter{UseSchema: true},
		},
		Outputs:               []kio.Writer{output},
		ContinueOnEmptyResult: true,
	}.Execute()
}

// Select keeps all of the input resources.
func (g *Generator) Select(nodes []*yaml.RNode) ([]*yaml.RNode, error) {
	// TODO Should we limit this only to types we can actually get information from?
	return nodes, nil
}

// Map scans for useful information to include in the application definition
func (g *Generator) Map(node *yaml.RNode, meta yaml.ResourceMeta) ([]interface{}, error) {
	var result []interface{}

	// If the resource stream already contains an application, we will use that
	// as a starting point for the rest of the generation.
	if meta.Kind == "Application" && meta.APIVersion == redskyappsv1alpha1.GroupVersion.String() {
		data, err := node.MarshalJSON()
		if err != nil {
			return nil, err
		}

		app := &redskyappsv1alpha1.Application{}
		if err := json.Unmarshal(data, app); err != nil {
			return nil, err
		}

		result = append(result, app)
	}

	return result, nil
}

// Transform converts the scan information into an application definition.
func (g *Generator) Transform(_ []*yaml.RNode, selected []interface{}) ([]*yaml.RNode, error) {
	result := scan.ObjectSlice{}

	app := &redskyappsv1alpha1.Application{}
	for _, sel := range selected {
		switch s := sel.(type) {

		case *redskyappsv1alpha1.Application:
			g.merge(s, app)

		}
	}

	g.apply(app)
	if err := g.clean(app); err != nil {
		return nil, err
	}

	result = append(result, app)
	return result.Read()
}

// merge a source application into another application.
func (g *Generator) merge(src, dst *redskyappsv1alpha1.Application) {
	if src.Name != "" {
		dst.Name = src.Name
	}

	if src.Namespace != "" {
		dst.Namespace = src.Namespace
	}

	if len(dst.Labels) > 0 {
		for k, v := range src.Labels {
			dst.Labels[k] = v
		}
	} else {
		dst.Labels = src.Labels
	}

	if len(dst.Annotations) > 0 {
		for k, v := range src.Annotations {
			dst.Annotations[k] = v
		}
	} else {
		dst.Annotations = src.Annotations
	}

	dst.Resources = append(dst.Resources, src.Resources...)
	dst.Scenarios = append(dst.Scenarios, src.Scenarios...)
	dst.Objectives = append(dst.Objectives, src.Objectives...)
}

// apply adds the generator configuration to the supplied application
func (g *Generator) apply(app *redskyappsv1alpha1.Application) {
	if g.Name != "" {
		app.Name = g.Name
	}

	if app.Annotations == nil {
		app.Annotations = make(map[string]string)
	}
	app.Annotations[redskyappsv1alpha1.AnnotationLastScanned] = time.Now().UTC().Format(time.RFC3339)

	app.Resources = append(app.Resources, g.Resources...)

	for _, o := range g.Objectives {
		app.Objectives = append(app.Objectives, redskyappsv1alpha1.Objective{Name: o})
	}
}

// clean ensures that the application state is reasonable.
func (g *Generator) clean(app *redskyappsv1alpha1.Application) error {
	resources := konjure.Resources{}
	for _, r := range app.Resources {
		if r.Resource != nil {
			var rstr []string
			for _, rr := range r.Resource.Resources {
				if rr != "-" && filepath.Dir(rr) != "/dev/fd" {
					rstr = append(rstr, rr)
				}
			}
			if len(rstr) == 0 {
				continue
			}
			r.Resource.Resources = rstr
		}

		resources = append(resources, r)
	}
	app.Resources = resources

	return nil
}