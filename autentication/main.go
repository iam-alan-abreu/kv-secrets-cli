package autentication


import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
)


func Autentication(vaultUrl string) (azcore.TokenCredential, error) {
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

func CheckAuthentication(client *azsecrets.Client) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	return checkAuthenticationWithContext(ctx, client)
}

func checkAuthenticationWithContext(ctx context.Context, client *azsecrets.Client) error {
	_, err := client.NewListSecretPropertiesPager(nil).NextPage(ctx)
	return err
}