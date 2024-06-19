package utils

import (
	"bufio"
	"os"
)

// addLineIfFileExists adiciona uma linha ao final de um arquivo se ele existir
func AddLineIfFileExists(filePath string, line string) error {
	// Verifica se o arquivo existe
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil // Se o arquivo não existe, apenas retorna
	}

	// Abre o arquivo no modo append e cria um writer
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	_, err = writer.WriteString(line + "\n")
	if err != nil {
		return err
	}
	return writer.Flush()
}
