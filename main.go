package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"cli-kv/autentication"
	"cli-kv/azcli"
	"cli-kv/secrets"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
)

func main() {
	// Definindo os parâmetros de entrada usando a biblioteca flag
	vaultName := flag.String("vaultName", "", "Name of the Azure Key Vault")
	appName := flag.String("appName", "", "Name of the application (optional)")
	outputPath := flag.String("outputPath", ".", "Path to save the .env file (optional)")
	subscription := flag.String("subscription", "", "Azure Subscription ID (optional)")
	loginParams := flag.String("loginParams", "", "Additional parameters for 'az login' (optional)")
	flag.Parse()

	// Verificando se o vaultName foi fornecido
	if *vaultName == "" {
		log.Fatalf("Usage: %s --vaultName <vaultName> [--appName <appName>] [--outputPath <outputPath>] [--subscription <subscription>] [--loginParams <loginParams>]", os.Args[0])
	}

	// Verifica se o usuário está logado
	if err := azcli.EnsureAzureLogin(*loginParams); err != nil {
		log.Fatalf("failed to ensure Azure login: %v", err)
	}

	// Configurar a assinatura do Azure, se fornecida
	if *subscription != "" {
		err := azcli.SetAzureSubscription(*subscription)
		if err != nil {
			log.Fatalf("failed to set Azure subscription: %v", err)
		}
	}

	outputFilename := filepath.Join(*outputPath, ".env")

	// Verifica se o diretório de saída existe, caso contrário, cria o diretório
	if _, err := os.Stat(*outputPath); os.IsNotExist(err) {
		err = os.MkdirAll(*outputPath, 0755)
		if err != nil {
			log.Fatalf("failed to create output directory: %v", err)
		}
	}

	// Verifica se o arquivo já existe
	if _, err := os.Stat(outputFilename); err == nil {
		log.Printf("Output file %s already exists. Exiting without doing anything.\n", outputFilename)
		return
	}

	vaultUrl := "https://" + *vaultName + ".vault.azure.net"

	cred, err := autentication.Autentication(vaultUrl)
	if err != nil {
		log.Fatalf("failed to obtain a credential: %v", err)
	}

	client, err := azsecrets.NewClient(vaultUrl, cred, nil)
	if err != nil {
		log.Fatalf("failed to create KeyVault client: %v", err)
	}

	// Verifica a autenticação tentando listar segredos.
	if err := autentication.CheckAuthentication(client); err != nil {
		log.Fatalf("authentication failed: %v", err)
	}

	// Obter segredos do Key Vault e salvar em um arquivo .env
	secrets.SaveSecretsToFile(client, *appName, outputFilename)
}


