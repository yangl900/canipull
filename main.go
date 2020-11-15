package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"

	"github.com/yangl900/canipull/pkg/authorizer"

	az "github.com/Azure/go-autorest/autorest/azure"
	"github.com/yangl900/canipull/pkg/exitcode"
	"github.com/yangl900/canipull/pkg/log"
	"k8s.io/legacy-cloud-providers/azure"

	flag "github.com/spf13/pflag"
)

const (
	DefaultAzureCfgPath string = "/etc/kubernetes/azure.json"
)

var (
	logLevel   *uint   = flag.UintP("verbose", "v", 2, "output verbosity level.")
	configPath *string = flag.String("config", "", "the azure.json config file path.")
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("No ACR input. Expect `canipull myacr.azurecr.io`.")
		return
	}

	flag.Parse()
	logger := log.NewLogger(*logLevel)

	acr := os.Args[1]
	if _, err := net.LookupHost(acr); err != nil {
		logger.V(2).Info("Checking host name resolution: FAILED")
		logger.V(2).Info("Failed to resolve specified fqdn %s: %s \n", acr, err)
		os.Exit(exitcode.DNSResolutionFailure)
	}
	logger.V(2).Info("Checking host name resolution: SUCCEEDED")

	azConfigPath := *configPath
	if *configPath == "" {
		azConfigPath = DefaultAzureCfgPath
	}

	logger.V(6).Info("Loading azure.json file from %s", azConfigPath)
	if _, err := os.Stat(azConfigPath); err != nil {
		logger.V(2).Info("Failed to load azure.json. Are you running inside Kubernetes on Azure? \n")
		os.Exit(exitcode.AzureConfigNotFound)
	}

	var cfg azure.Config
	configBytes, err := ioutil.ReadFile(azConfigPath)
	if err != nil {
		logger.V(2).Info("Failed to read azure.json file: %s \n", err)
		os.Exit(exitcode.AzureConfigReadFailure)
	}

	if err := json.Unmarshal(configBytes, &cfg); err != nil {
		logger.V(2).Info("Failed to read azure.json file: %s", err)
		os.Exit(exitcode.AzureConfigUnmarshalFailure)
	}

	if cfg.AADClientID == "msi" && cfg.AADClientSecret == "msi" {
		logger.V(2).Info("Checking managed identity...")
		os.Exit(validateMsiAuth(acr, cfg, logger))
		return
	}

	logger.V(4).Info("The cluster uses service principal.")
	os.Exit(validateServicePrincipalAuth(acr, cfg, logger))
}

func validateMsiAuth(acr string, cfg azure.Config, logger *log.Logger) int {
	env, err := az.EnvironmentFromName(cfg.Cloud)
	if err != nil {
		logger.V(2).Info("Unknown Azure cloud name: %s", cfg.Cloud)
		return exitcode.AzureCloudUnknown
	}

	tr := authorizer.NewTokenRetriever(env.ActiveDirectoryEndpoint)
	token, err := tr.AcquireARMToken(cfg.AADClientID)
	if err != nil {
		logger.V(2).Info("Validating managed identity existance: FAILED")
		logger.V(2).Info("Getting managed identity token failed with: %s", err)
		return exitcode.ServicePrincipalCredentialInvalid
	}
	logger.V(2).Info("Validating managed identity existance: SUCCEEDED")
	logger.V(6).Info("ARM access token: %s", token)

	te := authorizer.NewTokenExchanger()
	acrToken, err := te.ExchangeACRAccessToken(token, acr)
	if err != nil {
		logger.V(2).Info("Validating image pull permission: FAILED")
		logger.V(2).Info("ACR %s rejected token exchange: %s", acr, err)
		return exitcode.MissingImagePullPermision
	}

	logger.V(2).Info("Validating image pull permission: SUCCEEDED")
	logger.V(6).Info("ACR access token: %s", acrToken)
	return 0
}

func validateServicePrincipalAuth(acr string, cfg azure.Config, logger *log.Logger) int {
	env, err := az.EnvironmentFromName(cfg.Cloud)
	if err != nil {
		logger.V(2).Info("Unknown Azure cloud name: %s", cfg.Cloud)
		return exitcode.AzureCloudUnknown
	}

	tr := authorizer.NewTokenRetriever(env.ActiveDirectoryEndpoint)
	token, err := tr.AcquireARMTokenSP(cfg.AADClientID, cfg.AADClientSecret, cfg.TenantID)
	if err != nil {
		logger.V(2).Info("Validating service principal credential: FAILED")
		logger.V(2).Info("Sign in to AAD failed with: %s", err)
		return exitcode.ServicePrincipalCredentialInvalid
	}
	logger.V(2).Info("Validating service principal credential: SUCCEEDED")
	logger.V(6).Info("ARM access token: %s", token)

	te := authorizer.NewTokenExchanger()
	acrToken, err := te.ExchangeACRAccessToken(token, acr)
	if err != nil {
		logger.V(2).Info("Validating image pull permission: FAILED")
		logger.V(2).Info("ACR %s rejected token exchange: %s", acr, err)
		return exitcode.MissingImagePullPermision
	}

	logger.V(2).Info("Validating image pull permission: SUCCEEDED")
	logger.V(6).Info("ACR access token: %s", acrToken)
	return 0
}
