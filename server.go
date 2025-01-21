// server.go
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type ExchangeRateResponse struct {
	USDBRL struct {
		Bid string `json:"bid"`
	} `json:"USDBRL"`
}

func main() {
	db, err := sql.Open("sqlite3", "./cotacoes.db")
	if err != nil {
		log.Fatalf("Erro ao abrir o banco de dados: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS cotacoes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		valor TEXT,
		data_hora DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		log.Fatalf("Erro ao criar tabela no banco de dados: %v", err)
	}

	http.HandleFunc("/cotacao", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		cotacao, err := fetchExchangeRate(ctx)
		if err != nil {
			http.Error(w, fmt.Sprintf("Erro ao obter a cotação: %v", err), http.StatusInternalServerError)
			return
		}

		err = saveExchangeRate(ctx, db, cotacao)
		if err != nil {
			http.Error(w, fmt.Sprintf("Erro ao salvar a cotação no banco de dados: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"bid": cotacao})
	})

	log.Println("Servidor iniciado na porta 8080")
	http.ListenAndServe(":8080", nil)
}

func fetchExchangeRate(ctx context.Context) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
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

	var exchangeRateResponse ExchangeRateResponse
	if err := json.NewDecoder(resp.Body).Decode(&exchangeRateResponse); err != nil {
		return "", fmt.Errorf("erro ao decodificar resposta: %v", err)
	}

	return exchangeRateResponse.USDBRL.Bid, nil
}

func saveExchangeRate(ctx context.Context, db *sql.DB, value string) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Millisecond)
	defer cancel()

	_, err := db.ExecContext(ctx, "INSERT INTO cotacoes (valor) VALUES (?)", value)
	if err != nil {
		return fmt.Errorf("erro ao salvar no banco de dados: %v", err)
	}

	return nil
}
