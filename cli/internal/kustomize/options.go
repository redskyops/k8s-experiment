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

package kustomize

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/thestormforge/optimize-controller/v2/config"
	"sigs.k8s.io/kustomize/api/filesys"
	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/api/types"
)

type Option func(*Kustomize) error

const (
	defaultNamespace = "stormforge-system"
	defaultImage     = "controller:latest"
)

// This will get overridden at build time with the appropriate version image.
var BuildImage = defaultImage

func defaultOptions() *Kustomize {
	fs := filesys.MakeFsInMemory()

	return &Kustomize{
		Base:       "/",
		fs:         fs,
		Kustomizer: krusty.MakeKustomizer(krusty.MakeDefaultOptions()),
		kustomize:  &types.Kustomization{},
	}
}

// WithResources updates the kustomization with the specified list of
// Assets and writes them to the in memory filesystem.
func WithResources(efs embed.FS) Option {
	return func(k *Kustomize) (err error) {
		// There's a chance this could be fragile, but since its only
		// intended use is for our installation, we'll give it a shot
		k.kustomize.Resources = []string{"default"}

		err = fs.WalkDir(efs, ".", func(path string, info os.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			b, err := fs.ReadFile(efs, path)
			if err != nil {
				return err
			}

			if err = k.fs.WriteFile(filepath.Join(k.Base, path), b); err != nil {
				return err
			}
			return nil
		})

		return err
	}
}

// WithPatches updates the kustomization with the specified list of
// Patches and writes them to the in memory filesystem.
func WithPatches(patches []types.Patch) Option {
	return func(k *Kustomize) (err error) {
		// Write out all assets to in memory filesystem
		for _, patch := range patches {
			k.kustomize.Patches = append(k.kustomize.Patches, patch)

			if patch.Path == "" {
				continue
			}

			// TODO I wonder if this even makes sense...
			// If we include the patch above ( which would include the patch bytes ) this shouldnt ever get used?
			if err = k.fs.WriteFile(filepath.Join(k.Base, patch.Path), []byte(patch.Patch)); err != nil {
				return err
			}

		}
		return nil
	}
}

// WithInstall initializes a kustomization with the bases of what we need
// to perform an install/init.
func WithInstall() Option {
	return func(k *Kustomize) error {
		k.kustomize = &types.Kustomization{
			Namespace: defaultNamespace,
			Images: []types.Image{
				{
					Name:    defaultImage,
					NewName: strings.Split(BuildImage, ":")[0],
					NewTag:  strings.Split(BuildImage, ":")[1],
				},
			},
		}

		// Pull in the default bundled resources
		if err := WithResources(config.Content)(k); err != nil {
			return err
		}

		return nil
	}
}

// WithNamespace sets the namespace attribute for the kustomization.
func WithNamespace(n string) Option {
	return func(k *Kustomize) error {
		k.kustomize.Namespace = n
		return nil
	}
}

// WithImage sets the image attribute for the kustomiztion.
func WithImage(i string) Option {
	return func(k *Kustomize) error {
		imageParts := strings.Split(i, ":")
		if len(imageParts) != 2 {
			return fmt.Errorf("invalid image specified %s", i)
		}

		k.kustomize.Images = append(k.kustomize.Images, types.Image{
			Name:    BuildImage,
			NewName: imageParts[0],
			NewTag:  imageParts[1],
		})
		return nil
	}
}

// WithLabels sets the common labels attribute for the kustomization.
func WithLabels(l map[string]string) Option {
	return func(k *Kustomize) error {
		// The schema for plugins are loosely defined, so we need to use a template
		labelTransformer := `
apiVersion: builtin
kind: LabelTransformer
metadata:
  name: metadata_labels
labels:
{{ range $label, $value := . }}
  {{ $label }}: {{ $value }}
{{ end }}
fieldSpecs:
  - kind: Deployment
    path: spec/template/metadata/labels
    create: true
  - path: metadata/labels
    create: true`

		t := template.Must(template.New("labelTransformer").Parse(labelTransformer))

		// Execute the template for each recipient.
		var b bytes.Buffer
		if err := t.Execute(&b, l); err != nil {
			return err
		}

		if err := k.fs.WriteFile(filepath.Join(k.Base, "labelTransformer.yaml"), b.Bytes()); err != nil {
			return err
		}

		k.kustomize.Transformers = append(k.kustomize.Transformers, "labelTransformer.yaml")

		return nil
	}
}

// WithAPI configures the controller to use the Optimize API.
// If true, the controller deployment is patched to pull environment variables from the secret.
func WithAPI(o bool) Option {
	return func(k *Kustomize) error {
		if !o {
			return nil
		}

		// Since we've already generated the base assets with the `optimize` prefix
		// all of these resources need to reference that
		controllerEnvPatch := []byte(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: optimize-controller-manager
  namespace: stormforge-system
spec:
  template:
    spec:
      containers:
      - name: manager
        envFrom:
        - secretRef:
            name: optimize-manager`)

		if err := k.fs.WriteFile(filepath.Join(k.Base, "manager_patch.yaml"), controllerEnvPatch); err != nil {
			return err
		}

		k.kustomize.PatchesStrategicMerge = append(k.kustomize.PatchesStrategicMerge, "manager_patch.yaml")

		return nil
	}
}

func WithImagePullPolicy(pullPolicy string) Option {
	return func(k *Kustomize) error {
		controllerPullPolicyPatch := []byte(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: optimize-controller-manager
  namespace: stormforge-system
spec:
  template:
    spec:
      containers:
      - name: manager
        imagePullPolicy: ` + pullPolicy)

		if err := k.fs.WriteFile(filepath.Join(k.Base, "pull_policy_patch.yaml"), controllerPullPolicyPatch); err != nil {
			return err
		}

		k.kustomize.PatchesStrategicMerge = append(k.kustomize.PatchesStrategicMerge, "pull_policy_patch.yaml")

		return nil
	}
}

func WithFS(fs filesys.FileSystem) Option {
	return func(k *Kustomize) (err error) {
		k.fs = fs
		k.Kustomizer = krusty.MakeKustomizer(krusty.MakeDefaultOptions())

		return nil
	}
}

func WithResourceNames(filenames []string) Option {
	return func(k *Kustomize) (err error) {
		k.kustomize.Resources = append(k.kustomize.Resources, filenames...)
		return nil
	}
}
