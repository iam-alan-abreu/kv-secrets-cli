package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
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
	if err := ensureAzureLogin(*loginParams); err != nil {
		log.Fatalf("failed to ensure Azure login: %v", err)
	}

	// Configurar a assinatura do Azure, se fornecida
	if *subscription != "" {
		err := setAzureSubscription(*subscription)
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

func ensureAzureLogin(loginParams string) error {
	cmd := exec.Command("az", "account", "list")
	output, err := cmd.CombinedOutput()
	if err != nil || strings.Contains(string(output), "Please run \"az login\" to access your accounts.") {
		log.Println("Você não está logado, executando 'az login' para efetuar o login...")
		loginCmdArgs := []string{"login"}
		if loginParams != "" {
			loginCmdArgs = append(loginCmdArgs, strings.Split(loginParams, " ")...)
		}
		loginCmd := exec.Command("az", loginCmdArgs...)
		loginCmd.Stdout = os.Stdout
		loginCmd.Stderr = os.Stderr
		if loginErr := loginCmd.Run(); loginErr != nil {
			return fmt.Errorf("failed to execute 'az login': %w", loginErr)
		}
		log.Println("Login successful.")
		// Verifica o login novamente
		cmd = exec.Command("az", "account", "list")
		output, err = cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to list Azure accounts after login: %s", string(output))
		}
	}
	return nil
}

func setAzureSubscription(subscriptionID string) error {
	log.Printf("Setting Azure subscription to %s\n", subscriptionID)
	cmd := exec.Command("az", "account", "set", "--subscription", subscriptionID)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to set Azure subscription: %s", string(output))
	}
	log.Printf("Azure subscription set to %s successfully\n", subscriptionID)
	return nil
}

func authenticate(vaultUrl string) (azcore.TokenCredential, error) {
	log.Println("Authenticating with Azure CLI credentials...")
	cred, err := azidentity.NewAzureCLICredential(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Azure CLI credential: %w", err)
	}

	// Aumenta o tempo de espera para a autenticação
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// Teste a credencial para garantir que ela funcione
	client, err := azsecrets.NewClient(vaultUrl, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create KeyVault client for authentication test: %w", err)
	}
	if err := checkAuthenticationWithContext(ctx, client); err != nil {
		return nil, fmt.Errorf("authentication test failed: %w", err)
	}

	log.Println("Authentication successful.")
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
			fmt.Fprintf(output, "%s=%s\n", envVarName, envVarValue)
		}
	}
}
