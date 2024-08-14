//go:build e2e
// +build e2e

/*
Copyright © 2023 - 2024 SUSE LLC

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

package suites

import (
	_ "embed"
)

var (
	//go:embed data/cluster-templates/docker-kubeadm.yaml
	CAPIDockerKubeadm []byte

	//go:embed data/cluster-templates/docker-rke2.yaml
	CAPIDockerRKE2 []byte

	//go:embed data/rancher/settings-patch.yaml
	RancherSettingsPatch []byte
)
