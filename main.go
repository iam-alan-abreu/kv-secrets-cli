package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	// Definindo os parâmetros de entrada usando a biblioteca flag
	vaultName := flag.String("vaultName", "", "Name of the Azure Key Vault")
	appName := flag.String("appName", "", "Name of the application (optional)")
	outputPath := flag.String("outputPath", ".", "Path to save the .env file (optional)")
	flag.Parse()

	// Verificando se o vaultName foi fornecido
	if *vaultName == "" {
		log.Fatalf("Usage: %s --vaultName <vaultName> [--appName <appName>] [--outputPath <outputPath>]", os.Args[0])
	}

	outputFilename := filepath.Join(*outputPath, ".env")
	//fmt.Println(outputFilename)
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

	cred, err := authenticate(vaultUrl)
	if err != nil {
		log.Fatalf("failed to obtain a credential: %v", err)
	}

	client, err := azsecrets.NewClient(vaultUrl, cred, nil)
	if err != nil {
		log.Fatalf("failed to create KeyVault client: %v", err)
	}

	// Verifica a autenticação tentando listar segredos.
	if err := checkAuthentication(client); err != nil {
		log.Fatalf("authentication failed: %v", err)
	}

	// Obter segredos do Key Vault e salvar em um arquivo .env
	saveSecretsToFile(client, *appName, outputFilename)
}

func authenticate(vaultUrl string) (azcore.TokenCredential, error) {
	cred, err := azidentity.NewDeviceCodeCredential(&azidentity.DeviceCodeCredentialOptions{
		UserPrompt: func(ctx context.Context, deviceCode azidentity.DeviceCodeMessage) error {
			fmt.Println(deviceCode.Message)
			return nil
		},
	})
	if err != nil {
		return nil, err
	}

	// Aumenta o tempo de espera para a autenticação
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// Teste a credencial para garantir que ela funcione
	client, err := azsecrets.NewClient(vaultUrl, cred, nil)
	if err != nil {
		return nil, err
	}
	if err := checkAuthenticationWithContext(ctx, client); err != nil {
		return nil, err
	}

	return cred, nil
}

func checkAuthentication(client *azsecrets.Client) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	return checkAuthenticationWithContext(ctx, client)
}

func checkAuthenticationWithContext(ctx context.Context, client *azsecrets.Client) error {
	_, err := client.NewListSecretPropertiesPager(nil).NextPage(ctx)
	return err
}

func saveSecretsToFile(client *azsecrets.Client, appName, outputFilename string) {
	output, err := os.Create(outputFilename)
	if err != nil {
		log.Fatalf("failed to create output file: %v", err)
	}
	defer output.Close()

	pager := client.NewListSecretPropertiesPager(nil)
	for pager.More() {
		page, err := pager.NextPage(context.Background())
		if err != nil {
			log.Fatalf("failed to list secrets: %v", err)
		}
		for _, secret := range page.Value {
			secretName := secret.ID.Name()
			// Se appName for fornecido, filtra segredos pelo prefixo appName
			if appName != "" {
				if strings.HasPrefix(secretName, appName+"-") {
					secretName = strings.TrimPrefix(secretName, appName+"-")
				} else {
					continue
				}
			}
			resp, err := client.GetSecret(context.Background(), secret.ID.Name(), "", nil)
			if err != nil {
				log.Fatalf("failed to get secret: %v", err)
			}
			envVarName := strings.ReplaceAll(secretName, "-", "_")
			envVarValue := *resp.Value
			//fmt.Println(output)
			fmt.Fprintf(output, "%s=%s\n", envVarName, envVarValue)
		}
	}
}
