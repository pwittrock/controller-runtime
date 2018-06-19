/*
Copyright 2018 The Kubernetes Authors.

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

package test

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/ghodss/yaml"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apiextensions-apiserver/test/integration/testserver"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

// InstallCRDs installs a collection of CRDs into a cluster by reading the crd yaml files from a directory
func InstallCRDs(config *rest.Config, path string) (int, error) {
	if path == "" {
		path = "."
	}

	// Read the CRD yamls
	crds, err := readCrdsFromFiles(path)
	if err != nil {
		return 0, err
	}

	// Get deps to create CRDs
	cs, err := clientset.NewForConfig(config)
	if err != nil {
		return 0, err
	}
	cp := dynamic.NewDynamicClientPool(config)

	// Create the CRDs
	for _, crd := range crds {
		if _, err := testserver.CreateNewCustomResourceDefinition(crd, cs, cp); err != nil {
			return 0, err
		}
	}

	return len(crds), nil
}

// readCrdsFromFiles reads the CRDs from files and Unmarshals them into structs
func readCrdsFromFiles(path string) ([]*v1beta1.CustomResourceDefinition, error) {
	// Read the CRD files
	var files []os.FileInfo
	var err error
	if files, err = ioutil.ReadDir(path); err != nil {
		return nil, err
	}

	crds := []*v1beta1.CustomResourceDefinition{}
	for _, file := range files {
		// Only parse json and yamls
		ext := filepath.Ext(file.Name())
		if ext != ".json" && ext != ".yaml" {
			continue
		}

		// Unmarshal the file into a struct
		b, err := ioutil.ReadFile(filepath.Join(path, file.Name()))
		if err != nil {
			return nil, err
		}
		crd := &v1beta1.CustomResourceDefinition{}
		if err = yaml.Unmarshal(b, crd); err != nil {
			return nil, err
		}

		// Not a CRD
		if crd.Spec.Names.Kind == "" || crd.Spec.Group == "" {
			continue
		}

		crds = append(crds, crd)
	}
	return crds, nil
}
