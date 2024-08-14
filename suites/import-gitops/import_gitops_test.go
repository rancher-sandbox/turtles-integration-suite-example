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

package import_gitops

import (
	_ "embed"

	. "github.com/onsi/ginkgo/v2"
	. "sigs.k8s.io/controller-runtime/pkg/envtest/komega"

	"k8s.io/utils/ptr"

	"github.com/rancher/cluster-api-provider-rke2/turtles-integration-suite-example/suites"
	"github.com/rancher/turtles/test/e2e"
	"github.com/rancher/turtles/test/e2e/specs"
)

var _ = Describe("[Docker] [Kubeadm] Create and delete CAPI cluster functionality should work with namespace auto-import", func() {
	BeforeEach(func() {
		SetClient(setupClusterResult.BootstrapClusterProxy.GetClient())
		SetContext(ctx)
	})

	specs.CreateMgmtV3UsingGitOpsSpec(ctx, func() specs.CreateMgmtV3UsingGitOpsSpecInput {
		return specs.CreateMgmtV3UsingGitOpsSpecInput{
			E2EConfig:                 e2eConfig,
			BootstrapClusterProxy:     setupClusterResult.BootstrapClusterProxy,
			ClusterctlConfigPath:      flagVals.ConfigPath,
			ClusterctlBinaryPath:      e2eConfig.GetVariable(e2e.ClusterctlBinaryPathVar),
			ArtifactFolder:            e2eConfig.GetVariable(e2e.ArtifactsFolderVar),
			ClusterTemplate:           suites.CAPIDockerKubeadm,
			ClusterName:               "clusterv3-docker-kubeadm",
			ControlPlaneMachineCount:  ptr.To[int](1),
			WorkerMachineCount:        ptr.To[int](1),
			GitAddr:                   giteaResult.GitAddress,
			GitAuthSecretName:         e2e.AuthSecretName,
			SkipCleanup:               false,
			SkipDeletionTest:          false,
			LabelNamespace:            true,
			TestClusterReimport:       true,
			RancherServerURL:          hostName,
			CAPIClusterCreateWaitName: "wait-rancher",
			DeleteClusterWaitName:     "wait-controllers",
		}
	})
})

var _ = Describe("[Docker] [RKE2] Create and delete CAPI cluster functionality should work with namespace auto-import", func() {
	BeforeEach(func() {
		SetClient(setupClusterResult.BootstrapClusterProxy.GetClient())
		SetContext(ctx)
	})

	specs.CreateMgmtV3UsingGitOpsSpec(ctx, func() specs.CreateMgmtV3UsingGitOpsSpecInput {
		return specs.CreateMgmtV3UsingGitOpsSpecInput{
			E2EConfig:                 e2eConfig,
			BootstrapClusterProxy:     setupClusterResult.BootstrapClusterProxy,
			ClusterctlConfigPath:      flagVals.ConfigPath,
			ClusterctlBinaryPath:      e2eConfig.GetVariable(e2e.ClusterctlBinaryPathVar),
			ArtifactFolder:            e2eConfig.GetVariable(e2e.ArtifactsFolderVar),
			ClusterTemplate:           e2e.CAPIDockerRKE2,
			ClusterName:               "clusterv3-docker-rke2",
			ControlPlaneMachineCount:  ptr.To(1),
			WorkerMachineCount:        ptr.To[int](1),
			GitAddr:                   giteaResult.GitAddress,
			GitAuthSecretName:         e2e.AuthSecretName,
			SkipCleanup:               false,
			SkipDeletionTest:          false,
			LabelNamespace:            true,
			TestClusterReimport:       true,
			RancherServerURL:          hostName,
			CAPIClusterCreateWaitName: "wait-rancher",
			DeleteClusterWaitName:     "wait-controllers",
		}
	})
})
