package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const dolarRealExchangeUrl = "https://economia.awesomeapi.com.br/json/last/USD-BRL"

type USDBRL struct {
	Bid string `json:"bid"`
}

type Exchange struct {
	USDBRL USDBRL `json:"USDBRL"`
}

type Cotacao struct {
	gorm.Model
	ID  int `gorm:"primaryKey"`
	BID string
}

func main() {
	fmt.Println("Server started")

	fmt.Println("Connecting to database")
	db, err := gorm.Open(sqlite.Open("cotacao.db"), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	db.AutoMigrate(&Cotacao{})

	http.HandleFunc("/cotacao", func(w http.ResponseWriter, r *http.Request) {
		cotacao, err := GetDolarRealExchange()
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		InsertCotacao(db, Cotacao{BID: cotacao.USDBRL.Bid})
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cotacao.USDBRL)

	})

	fmt.Println("Server running on port 8080")
	http.ListenAndServe(":8080", nil)

}

func GetDolarRealExchange() (Exchange, error) {
	fmt.Println("Getting dolar real exchange")
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", dolarRealExchangeUrl, nil)
	if err != nil {
		return Exchange{}, err
	}

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return Exchange{}, err
	}

	defer resp.Body.Close()

	var cotacao Exchange
	err = json.NewDecoder(resp.Body).Decode(&cotacao)
	if err != nil {
		return Exchange{}, err
	}

	return cotacao, nil

}

func InsertCotacao(db *gorm.DB, cotacao Cotacao) error {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Millisecond)
	defer cancel()
	db.WithContext(ctx).Create(&cotacao)
	return nil
}
