package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type BrasilAPIResponse struct {
	Cep          string `json:"cep"`
	State        string `json:"state"`
	City         string `json:"city"`
	Neighborhood string `json:"neighborhood"`
	Street       string `json:"street"`
	Service      string `json:"service"`
}

type ViaCEPResponse struct {
	Cep         string `json:"cep"`
	Logradouro  string `json:"logradouro"`
	Complemento string `json:"complemento"`
	Bairro      string `json:"bairro"`
	Localidade  string `json:"localidade"`
	Uf          string `json:"uf"`
	Ibge        string `json:"ibge"`
	Gia         string `json:"gia"`
	Ddd         string `json:"ddd"`
	Siafi       string `json:"siafi"`
}

type CepResponse struct {
	Cep        string
	Logradouro string
	Bairro     string
	Localidade string
	Uf         string
	ApiSource  string
}

func buscaBrasilAPI(cep string) (*CepResponse, error) {

	client := &http.Client{
		Timeout: time.Second * 1,
	}

	resp, err := client.Get(fmt.Sprintf("https://brasilapi.com.br/api/cep/v1/%s", cep))
	if err != nil {
		return nil, fmt.Errorf("erro ao fazer requisição: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler corpo da resposta: %v", err)
	}

	var brasilAPIResponse BrasilAPIResponse

	err = json.Unmarshal(body, &brasilAPIResponse)
	if err != nil {
		return nil, fmt.Errorf("erro ao decodificar JSON: %v", err)
	}

	cepResponse := &CepResponse{
		Cep:        brasilAPIResponse.Cep,
		Logradouro: brasilAPIResponse.Street,
		Bairro:     brasilAPIResponse.Neighborhood,
		Localidade: brasilAPIResponse.City,
		Uf:         brasilAPIResponse.State,
		ApiSource:  "BrasilAPI",
	}

	return cepResponse, nil
}

func buscaViaCEP(cep string) (*CepResponse, error) {
	client := &http.Client{
		Timeout: time.Second * 1,
	}

	resp, err := client.Get(fmt.Sprintf("https://viacep.com.br/ws/%s/json/", cep))
	if err != nil {
		return nil, fmt.Errorf("erro ao fazer requisição: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler corpo da resposta: %v", err)
	}

	var viaCEPResponse ViaCEPResponse

	err = json.Unmarshal(body, &viaCEPResponse)
	if err != nil {
		return nil, fmt.Errorf("erro ao decodificar JSON: %v", err)
	}

	cepResponse := &CepResponse{
		Cep:        viaCEPResponse.Cep,
		Logradouro: viaCEPResponse.Logradouro,
		Bairro:     viaCEPResponse.Bairro,
		Localidade: viaCEPResponse.Localidade,
		Uf:         viaCEPResponse.Uf,
		ApiSource:  "ViaCEP",
	}

	return cepResponse, nil
}

func main() {
	cep := "01001000"

	canalBrasilAPI := make(chan *CepResponse, 1)
	canalViaCEP := make(chan *CepResponse, 1)
	canalErro := make(chan error, 1)

	go func() {
		resultado, err := buscaBrasilAPI(cep)
		if err != nil {
			canalErro <- fmt.Errorf("erro na BrasilAPI: %v", err)
			return
		}
		canalBrasilAPI <- resultado
	}()

	go func() {
		resultado, err := buscaViaCEP(cep)
		if err != nil {
			canalErro <- fmt.Errorf("erro no ViaCEP: %v", err)
			return
		}
		canalViaCEP <- resultado
	}()

	select {
	case resultado := <-canalViaCEP:
		fmt.Printf("ViaCEP foi mais rápido!\n")
		fmt.Printf("Resultado: %+v\n", resultado)
	case resultado := <-canalBrasilAPI:
		fmt.Printf("BrasilAPI foi mais rápida!\n")
		fmt.Printf("Resultado: %+v\n", resultado)
	case err := <-canalErro:
		fmt.Printf("Erro: %v\n", err)
	case <-time.After(time.Second):
		fmt.Println("Timeout: Nenhuma API respondeu em até 1 segundo")
	}
}
