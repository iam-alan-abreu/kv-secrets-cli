package secrets

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
)


func SaveSecretsToFile(client *azsecrets.Client, appName, outputFilename string) {
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