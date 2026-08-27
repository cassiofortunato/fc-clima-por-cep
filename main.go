package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

// Temperaturas representa o contrato de resposta de sucesso da API.
type Temperaturas struct {
	TempC float64 `json:"temp_C"`
	TempF float64 `json:"temp_F"`
	TempK float64 `json:"temp_K"`
}

type respostaViaCEP struct {
	Localidade string `json:"localidade"`
}

type respostaWeatherAPI struct {
	Current struct {
		TempC float64 `json:"temp_c"`
	} `json:"current"`
}

var (
	urlViaCEP  = "https://viacep.com.br/ws/%s/json/"
	urlWeather = "https://api.weatherapi.com/v1/current.json"

	regexCEP = regexp.MustCompile(`^\d{8}$`)

	errCEPNaoEncontrado = errors.New("can not find zipcode")
)

func cepValido(cep string) bool {
	return regexCEP.MatchString(cep)
}

func celsiusParaFahrenheit(celsius float64) float64 {
	return celsius*1.8 + 32
}

func celsiusParaKelvin(celsius float64) float64 {
	return celsius + 273
}

// buscarLocalidade consulta o ViaCEP e devolve o nome da cidade do CEP informado.
func buscarLocalidade(cep string) (string, error) {
	resp, err := http.Get(fmt.Sprintf(urlViaCEP, cep))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errCEPNaoEncontrado
	}

	var dados respostaViaCEP
	if err := json.NewDecoder(resp.Body).Decode(&dados); err != nil {
		return "", err
	}

	if dados.Localidade == "" {
		return "", errCEPNaoEncontrado
	}

	return dados.Localidade, nil
}

// buscarTemperatura consulta a WeatherAPI e devolve a temperatura atual em Celsius.
func buscarTemperatura(cidade string) (float64, error) {
	req, err := http.NewRequest(http.MethodGet, urlWeather, nil)
	if err != nil {
		return 0, err
	}

	parametros := req.URL.Query()
	parametros.Add("key", os.Getenv("WEATHER_API_KEY"))
	parametros.Add("q", cidade)
	req.URL.RawQuery = parametros.Encode()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("erro ao consultar clima: status %d", resp.StatusCode)
	}

	var dados respostaWeatherAPI
	if err := json.NewDecoder(resp.Body).Decode(&dados); err != nil {
		return 0, err
	}

	return dados.Current.TempC, nil
}

func manipuladorClima(w http.ResponseWriter, r *http.Request) {
	cep := strings.TrimPrefix(r.URL.Path, "/")

	if !cepValido(cep) {
		http.Error(w, "invalid zipcode", http.StatusUnprocessableEntity)
		return
	}

	cidade, err := buscarLocalidade(cep)
	if err != nil {
		if errors.Is(err, errCEPNaoEncontrado) {
			http.Error(w, "can not find zipcode", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tempC, err := buscarTemperatura(cidade)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resposta := Temperaturas{
		TempC: tempC,
		TempF: celsiusParaFahrenheit(tempC),
		TempK: celsiusParaKelvin(tempC),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resposta)
}

func main() {
	http.HandleFunc("/", manipuladorClima)

	porta := os.Getenv("PORT")
	if porta == "" {
		porta = "8080"
	}

	log.Printf("servidor rodando na porta %s", porta)
	log.Fatal(http.ListenAndServe(":"+porta, nil))
}
