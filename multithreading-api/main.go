package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

func requestBrasilAPI(cep string, ch chan<- map[string]string, errCh chan<- error) {
	url := fmt.Sprintf("https://brasilapi.com.br/api/cep/v1/%s", cep)
	resp, err := http.Get(url)
	if err != nil {
		errCh <- err
		return
	}
	defer resp.Body.Close()

	var address map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&address); err != nil {
		errCh <- err
		return
	}
	address["api_datasource"] = "BrasilAPI"

	ch <- address
}

func requestViaCEP(cep string, ch chan<- map[string]string, errCh chan<- error) {
	url := fmt.Sprintf("http://viacep.com.br/ws/%s/json/", cep)
	resp, err := http.Get(url)
	if err != nil {
		errCh <- err
		return
	}
	defer resp.Body.Close()

	var address map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&address); err != nil {
		errCh <- err
		return
	}
	address["api_datasource"] = "ViaCEP"

	ch <- address
}

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/ceps/{cep}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)

        var cep string
		cep = vars["cep"]
        if len(cep) != 8 {
            http.Error(w, "CEP inválido", http.StatusBadRequest)
            return
        }

		ch := make(chan map[string]string)
		errCh := make(chan error)

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		go requestBrasilAPI(cep, ch, errCh)
		go requestViaCEP(cep, ch, errCh)

		select {
		case address := <-ch:
			fmt.Printf("Endereço recebido: %+v\n", address)
			cancel()
		case err := <-errCh:
			fmt.Println("Erro ao buscar endereço:", err)
		case <- ctx.Done():
			fmt.Println("Timeout: Nenhuma API respondeu a tempo.")
		}
	})

	http.ListenAndServe(":8080", r)
}
