// client.go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	cotacao, err := getExchangeRate(ctx)
	if err != nil {
		log.Fatalf("Erro ao obter a cotação: %v", err)
	}

	err = saveToFile(cotacao)
	if err != nil {
		log.Fatalf("Erro ao salvar a cotação no arquivo: %v", err)
	}

	log.Println("Cotação salva com sucesso em cotacao.txt")
}

func getExchangeRate(ctx context.Context) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 300*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:8080/cotacao", nil)
	if err != nil {
		return "", fmt.Errorf("erro ao criar requisição: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("erro ao fazer requisição: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status não esperado: %v", resp.StatusCode)
	}

	var response map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("erro ao decodificar resposta: %v", err)
	}

	cotacao, ok := response["bid"]
	if !ok {
		return "", fmt.Errorf("campo 'bid' não encontrado na resposta")
	}

	return cotacao, nil
}

func saveToFile(cotacao string) error {
	file, err := os.Create("cotacao.txt")
	if err != nil {
		return fmt.Errorf("erro ao criar arquivo: %v", err)
	}
	defer file.Close()

	_, err = file.WriteString(fmt.Sprintf("Dólar: %s", cotacao))
	if err != nil {
		return fmt.Errorf("erro ao escrever no arquivo: %v", err)
	}

	return nil
}
