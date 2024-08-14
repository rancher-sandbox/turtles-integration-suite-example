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
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rancher/cluster-api-provider-rke2/turtles-integration-suite-example/suites"
	"github.com/rancher/turtles/test/e2e"
	turtlesframework "github.com/rancher/turtles/test/framework"
	"github.com/rancher/turtles/test/testenv"
	"k8s.io/klog/v2"
	"sigs.k8s.io/cluster-api/test/framework/clusterctl"
	ctrl "sigs.k8s.io/controller-runtime"
)

// Test suite flags.
var (
	flagVals *e2e.FlagValues
)

// Test suite global vars.
var (
	// e2eConfig to be used for this test, read from configPath.
	e2eConfig *clusterctl.E2EConfig

	// clusterctlConfigPath to be used for this test, created by generating a clusterctl local repository
	// with the providers specified in the configPath.
	clusterctlConfigPath string

	// hostName is the host name for the Rancher Manager server.
	hostName string

	ctx = context.Background()

	// setupClusterResult to be used for this test, created by setting up a management cluster.
	setupClusterResult *testenv.SetupTestClusterResult

	// giteaResult to be used for this test, created by deploying Gitea.
	giteaResult *testenv.DeployGiteaResult
)

func init() {
	// Initialize the test suite flags. Currently, only the configPath flag is used.
	flagVals = &e2e.FlagValues{}
	e2e.InitFlags(flagVals)
}

func TestE2E(t *testing.T) {
	RegisterFailHandler(Fail)

	ctrl.SetLogger(klog.Background())

	RunSpecs(t, "rancher-turtles-e2e-managementv3")
}

var _ = BeforeSuite(func() {
	Expect(flagVals.ConfigPath).To(BeAnExistingFile(), "Invalid test suite argument. e2e.config should be an existing file.")

	By(fmt.Sprintf("Loading the e2e test configuration from %q", flagVals.ConfigPath))
	e2eConfig = e2e.LoadE2EConfig(flagVals.ConfigPath)

	preSetupOutput := testenv.PreManagementClusterSetupHook(e2eConfig) // This hook is optional use only if management cluster environment matches defaults.

	By(fmt.Sprintf("Creating a clusterctl config into %q", e2eConfig.GetVariable(e2e.ArtifactsFolderVar)))
	clusterctlConfigPath = e2e.CreateClusterctlLocalRepository(ctx, e2eConfig, filepath.Join(e2eConfig.GetVariable(e2e.ArtifactsFolderVar), "repository"))

	useExistingCluter, err := strconv.ParseBool(e2eConfig.GetVariable(e2e.UseExistingClusterVar))
	Expect(err).ToNot(HaveOccurred(), "Invalid test suite argument. Can't parse e2e.use-existing-cluster %q", e2eConfig.GetVariable(e2e.UseExistingClusterVar))

	// Set up the management cluster, depending on the useExistingCluster value cluster will be created or used. If custom cluster provider is set, it will be used.
	// Currently, there are 2 built-in options "kind" and "EKS".
	setupClusterResult = testenv.SetupTestCluster(ctx, testenv.SetupTestClusterInput{
		UseExistingCluster:    useExistingCluter,
		E2EConfig:             e2eConfig,
		ClusterctlConfigPath:  clusterctlConfigPath,
		Scheme:                e2e.InitScheme(),
		ArtifactFolder:        e2eConfig.GetVariable(e2e.ArtifactsFolderVar),
		KubernetesVersion:     e2eConfig.GetVariable(e2e.KubernetesManagementVersionVar),
		HelmBinaryPath:        e2eConfig.GetVariable(e2e.HelmBinaryPathVar),
		CustomClusterProvider: preSetupOutput.CustomClusterProvider,
	})

	// Deploy ingress used by Rancher. There 3 options: EKS nginx, Ngrok, and Custom Ingress. See RancherDeployIngress documentation for more details.
	testenv.RancherDeployIngress(ctx, testenv.RancherDeployIngressInput{
		BootstrapClusterProxy:    setupClusterResult.BootstrapClusterProxy,
		HelmBinaryPath:           e2eConfig.GetVariable(e2e.HelmBinaryPathVar),
		HelmExtraValuesPath:      filepath.Join(e2eConfig.GetVariable(e2e.HelmExtraValuesFolderVar), "deploy-rancher-ingress.yaml"),
		IngressType:              preSetupOutput.IngressType,
		CustomIngress:            e2e.NginxIngress,
		CustomIngressNamespace:   e2e.NginxIngressNamespace,
		CustomIngressDeployment:  e2e.NginxIngressDeployment,
		IngressWaitInterval:      e2eConfig.GetIntervals(setupClusterResult.BootstrapClusterProxy.GetName(), "wait-rancher"),
		NgrokApiKey:              e2eConfig.GetVariable(e2e.NgrokApiKeyVar),
		NgrokAuthToken:           e2eConfig.GetVariable(e2e.NgrokAuthTokenVar),
		NgrokPath:                e2eConfig.GetVariable(e2e.NgrokPathVar),
		NgrokRepoName:            e2eConfig.GetVariable(e2e.NgrokRepoNameVar),
		NgrokRepoURL:             e2eConfig.GetVariable(e2e.NgrokUrlVar),
		DefaultIngressClassPatch: e2e.IngressClassPatch,
	})

	// Rancher deployment input. Every fiels is documented in the RancherDeployInput struct.
	rancherInput := testenv.DeployRancherInput{
		BootstrapClusterProxy:  setupClusterResult.BootstrapClusterProxy,
		HelmBinaryPath:         e2eConfig.GetVariable(e2e.HelmBinaryPathVar),
		HelmExtraValuesPath:    filepath.Join(e2eConfig.GetVariable(e2e.HelmExtraValuesFolderVar), "deploy-rancher.yaml"),
		InstallCertManager:     true,
		CertManagerChartPath:   e2eConfig.GetVariable(e2e.CertManagerPathVar),
		CertManagerUrl:         e2eConfig.GetVariable(e2e.CertManagerUrlVar),
		CertManagerRepoName:    e2eConfig.GetVariable(e2e.CertManagerRepoNameVar),
		RancherChartRepoName:   e2eConfig.GetVariable(e2e.RancherRepoNameVar),
		RancherChartURL:        e2eConfig.GetVariable(e2e.RancherUrlVar),
		RancherChartPath:       e2eConfig.GetVariable(e2e.RancherPathVar),
		RancherVersion:         e2eConfig.GetVariable(e2e.RancherVersionVar),
		RancherNamespace:       e2e.RancherNamespace,
		RancherPassword:        e2eConfig.GetVariable(e2e.RancherPasswordVar),
		RancherPatches:         [][]byte{e2e.RancherSettingPatch, suites.RancherSettingsPatch},
		RancherWaitInterval:    e2eConfig.GetIntervals(setupClusterResult.BootstrapClusterProxy.GetName(), "wait-rancher"),
		ControllerWaitInterval: e2eConfig.GetIntervals(setupClusterResult.BootstrapClusterProxy.GetName(), "wait-controllers"),
		Variables:              e2eConfig.Variables,
	}

	// Rancher pre-install hook. This hook is optional use only if management cluster environment matches defaults.
	rancherHookResult := testenv.PreRancherInstallHook(
		&testenv.PreRancherInstallHookInput{
			Ctx:                ctx,
			RancherInput:       &rancherInput,
			E2EConfig:          e2eConfig,
			SetupClusterResult: setupClusterResult,
			PreSetupOutput:     preSetupOutput,
		})

	hostName = rancherHookResult.HostName // Setting the host name for the Rancher Manager server is required. In this this done based on the Rancher pre-install hook output.

	// Deploy Rancher. This function will deploy Rancher and wait for it to be ready. More details can be found in the RancherDeploy documentation.
	testenv.DeployRancher(ctx, rancherInput)

	// Deploy Rancher Turtles. This function will deploy Rancher Turtles and wait for it to be ready. More details can be found in the DeployRancherTurtles documentation.
	rtInput := testenv.DeployRancherTurtlesInput{
		BootstrapClusterProxy:        setupClusterResult.BootstrapClusterProxy,
		HelmBinaryPath:               e2eConfig.GetVariable(e2e.HelmBinaryPathVar),
		TurtlesChartRepoName:         e2eConfig.GetVariable(e2e.TurtlesRepoNameVar),
		TurtlesChartPath:             e2eConfig.GetVariable(e2e.TurtlesPathVar),
		TurtlesChartUrl:              e2eConfig.GetVariable(e2e.TurtlesUrlVar),
		CAPIProvidersYAML:            e2e.CapiProviders,
		Namespace:                    turtlesframework.DefaultRancherTurtlesNamespace,
		WaitDeploymentsReadyInterval: e2eConfig.GetIntervals(setupClusterResult.BootstrapClusterProxy.GetName(), "wait-controllers"),
		Version:                      e2eConfig.GetVariable(e2e.TurtlesVersionVar),
		AdditionalValues: map[string]string{
			"rancherTurtles.features.addon-provider-fleet.enabled": "true", // Enable the Fleet addon provider. This is an example of how to enable a feature.
		},
	}

	// Deploy Rancher Turtles. This function will deploy Rancher Turtles and wait for it to be ready. More details can be found in the DeployRancherTurtles documentation.
	testenv.DeployRancherTurtles(ctx, rtInput)

	// Deploy Gitea. This function will deploy Gitea and wait for it to be ready. More details can be found in the DeployGitea documentation.
	giteaInput := testenv.DeployGiteaInput{
		BootstrapClusterProxy: setupClusterResult.BootstrapClusterProxy,
		HelmBinaryPath:        e2eConfig.GetVariable(e2e.HelmBinaryPathVar),
		ChartRepoName:         e2eConfig.GetVariable(e2e.GiteaRepoNameVar),
		ChartRepoURL:          e2eConfig.GetVariable(e2e.GiteaRepoURLVar),
		ChartName:             e2eConfig.GetVariable(e2e.GiteaChartNameVar),
		ChartVersion:          e2eConfig.GetVariable(e2e.GiteaChartVersionVar),
		ValuesFilePath:        "../data/gitea/values.yaml",
		Values: map[string]string{
			"gitea.admin.username": e2eConfig.GetVariable(e2e.GiteaUserNameVar),
			"gitea.admin.password": e2eConfig.GetVariable(e2e.GiteaUserPasswordVar),
		},
		RolloutWaitInterval: e2eConfig.GetIntervals(setupClusterResult.BootstrapClusterProxy.GetName(), "wait-gitea"),
		ServiceWaitInterval: e2eConfig.GetIntervals(setupClusterResult.BootstrapClusterProxy.GetName(), "wait-gitea-service"),
		AuthSecretName:      e2e.AuthSecretName,
		Username:            e2eConfig.GetVariable(e2e.GiteaUserNameVar),
		Password:            e2eConfig.GetVariable(e2e.GiteaUserPasswordVar),
		CustomIngressConfig: e2e.GiteaIngress, // This is an example of how to use a custom ingress configuration.
		Variables:           e2eConfig.Variables,
	}

	// This hook is optional use only if management cluster environment matches defaults. The one used below will set the service type based on the management cluster provider.
	testenv.PreGiteaInstallHook(&giteaInput, e2eConfig)

	// Deploy Gitea. This function will deploy Gitea and wait for it to be ready. More details can be found in the DeployGitea documentation.
	giteaResult = testenv.DeployGitea(ctx, giteaInput)
})

var _ = AfterSuite(func() {
	testenv.UninstallGitea(ctx, testenv.UninstallGiteaInput{
		BootstrapClusterProxy: setupClusterResult.BootstrapClusterProxy,
		HelmBinaryPath:        e2eConfig.GetVariable(e2e.HelmBinaryPathVar),
		DeleteWaitInterval:    e2eConfig.GetIntervals(setupClusterResult.BootstrapClusterProxy.GetName(), "wait-gitea-uninstall"),
	})

	skipCleanup, err := strconv.ParseBool(e2eConfig.GetVariable(e2e.SkipResourceCleanupVar))
	Expect(err).ToNot(HaveOccurred(), "Invalid test suite argument. Can't parse e2e.skip-cleanup %q", e2eConfig.GetVariable(e2e.SkipResourceCleanupVar))

	testenv.CleanupTestCluster(ctx, testenv.CleanupTestClusterInput{
		SetupTestClusterResult: *setupClusterResult,
		SkipCleanup:            skipCleanup,
		ArtifactFolder:         e2eConfig.GetVariable(e2e.ArtifactsFolderVar),
	})
})

func validateConfigVariables() {
	Expect(os.MkdirAll(e2eConfig.GetVariable(e2e.ArtifactsFolderVar), 0o755)).To(Succeed(), "Invalid test suite argument. Can't create e2e.artifacts-folder %q", e2eConfig.GetVariable(e2e.ArtifactsFolderVar))
	Expect(e2eConfig.GetVariable(e2e.ArtifactsFolderVar)).To(BeAnExistingFile(), "Invalid test suite argument. helm-binary-path should be an existing file.")
}
