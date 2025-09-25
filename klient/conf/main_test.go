/*
Copyright 2021 The Kubernetes Authors.

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

package conf

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	log "k8s.io/klog/v2"

	"k8s.io/client-go/util/homedir"
)

var (
	kubeconfigpath string
	kubeContext    string
)

func TestMain(m *testing.M) {
	setup()
	setKubeconfigFlags()
	code := m.Run()
	teardown()
	os.Exit(code)
}

func setup() {
	home := homedir.HomeDir()

	kubeconfigdir := filepath.Join(home, "test", ".kube")
	kubeconfigpath = filepath.Join(kubeconfigdir, "config")
	kubeContext = "test-context"

	// check if file exists
	_, err := os.Stat(kubeconfigpath)
	// create file if not exists
	if os.IsNotExist(err) {
		err = os.MkdirAll(kubeconfigdir, 0o777)
		if err != nil {
			log.ErrorS(err, "failed to create .kube dir")
			return
		}

		// generate kube config data
		data := genKubeconfig(kubeContext)

		err = createFile(kubeconfigpath, data)
		if err != nil {
			log.ErrorS(err, "failed to create config file")
			return
		}
	}

	log.Info("file created successfully", kubeconfigpath)

	flag.StringVar(&kubeconfig, "kubeconfig", "", "Paths to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&kubeContext, "context", "", "The name of the kubeconfig context to use. Only required if out-of-cluster.")
}

func setKubeconfigFlags() {
	// set --kubeconfig flag
	if err := flag.Set("kubeconfig", kubeconfigpath); err != nil {
		log.ErrorS(err, "unexpected error while setting kubeconfig flag value")
		return
	}

	// set --context flag
	if err := flag.Set("context", kubeContext); err != nil {
		log.ErrorS(err, "unexpected error while setting context flag value")
		return
	}

	flag.Parse()
}

func clearKubeconfigFlags() {
	// clear --kubeconfig flag
	if err := flag.Set("kubeconfig", ""); err != nil {
		log.ErrorS(err, "unexpected error while setting kubeconfig flag value")
		return
	}

	// clear --context flag
	if err := flag.Set("context", ""); err != nil {
		log.ErrorS(err, "unexpected error while setting context flag value")
		return
	}

	flag.Parse()
}

func createFile(path, data string) error {
	return os.WriteFile(path, []byte(data), 0o644)
}

// genKubeconfig used to genearte kube config file
// we can provide multiple contexts as well
func genKubeconfig(contexts ...string) string {
	var sb strings.Builder
	sb.WriteString(`---
apiVersion: v1
kind: Config
clusters:
`)
	for _, ctx := range contexts {
		sb.WriteString(`- cluster:
    server: ` + ctx + `
  name: ` + ctx + `
`)
	}
	sb.WriteString("contexts:\n")
	for _, ctx := range contexts {
		sb.WriteString(`- context:
    cluster: ` + ctx + `
    user: ` + ctx + `
  name: ` + ctx + `
`)
	}

	sb.WriteString("users:\n")
	for _, ctx := range contexts {
		sb.WriteString(`- name: ` + ctx + `
`)
	}
	sb.WriteString("preferences: {}\n")
	if len(contexts) > 0 {
		sb.WriteString("current-context: " + contexts[0] + "\n")
	}

	return sb.String()
}

func teardown() {
	home := homedir.HomeDir()
	err := os.RemoveAll(filepath.Join(home, "test"))
	if err != nil {
		log.ErrorS(err, "failed to delete .kube dir")
	}
}
