package azcli

import (

	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"


)
func EnsureAzureLogin(loginParams string) error {
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

func SetAzureSubscription(subscriptionID string) error {
	log.Printf("Setting Azure subscription to %s\n", subscriptionID)
	cmd := exec.Command("az", "account", "set", "--subscription", subscriptionID)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to set Azure subscription: %s", string(output))
	}
	log.Printf("Azure subscription set to %s successfully\n", subscriptionID)
	return nil
}
