package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func aproximado(a, b float64) bool {
	diff := a - b
	if diff < 0 {
		diff = -diff
	}
	return diff < 0.01
}

func TestCelsiusParaFahrenheit(t *testing.T) {
	casos := []struct {
		celsius  float64
		esperado float64
	}{
		{0, 32},
		{28.5, 83.3},
		{100, 212},
	}

	for _, c := range casos {
		if got := celsiusParaFahrenheit(c.celsius); !aproximado(got, c.esperado) {
			t.Errorf("celsiusParaFahrenheit(%.1f) = %.2f; esperado %.2f", c.celsius, got, c.esperado)
		}
	}
}

func TestCelsiusParaKelvin(t *testing.T) {
	casos := []struct {
		celsius  float64
		esperado float64
	}{
		{0, 273},
		{28.5, 301.5},
		{100, 373},
	}

	for _, c := range casos {
		if got := celsiusParaKelvin(c.celsius); !aproximado(got, c.esperado) {
			t.Errorf("celsiusParaKelvin(%.1f) = %.2f; esperado %.2f", c.celsius, got, c.esperado)
		}
	}
}

func TestCEPValido(t *testing.T) {
	casos := []struct {
		cep      string
		esperado bool
	}{
		{"01001000", true},
		{"1234567", false},
		{"123456789", false},
		{"0100100a", false},
		{"01001-000", false},
		{"", false},
	}

	for _, c := range casos {
		if got := cepValido(c.cep); got != c.esperado {
			t.Errorf("cepValido(%q) = %v; esperado %v", c.cep, got, c.esperado)
		}
	}
}

func TestManipuladorCEPInvalido(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/123", nil)
	rec := httptest.NewRecorder()

	manipuladorClima(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d; esperado %d", rec.Code, http.StatusUnprocessableEntity)
	}
}

func TestManipuladorCEPNaoEncontrado(t *testing.T) {
	servidorViaCEP := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"erro": "true"}`))
	}))
	defer servidorViaCEP.Close()

	urlOriginal := urlViaCEP
	urlViaCEP = servidorViaCEP.URL + "/%s"
	defer func() { urlViaCEP = urlOriginal }()

	req := httptest.NewRequest(http.MethodGet, "/99999999", nil)
	rec := httptest.NewRecorder()

	manipuladorClima(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d; esperado %d", rec.Code, http.StatusNotFound)
	}
}

func TestManipuladorSucesso(t *testing.T) {
	servidorViaCEP := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"localidade": "São Paulo"}`))
	}))
	defer servidorViaCEP.Close()

	servidorWeather := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"current": {"temp_c": 28.5}}`))
	}))
	defer servidorWeather.Close()

	urlViaCEPOriginal, urlWeatherOriginal := urlViaCEP, urlWeather
	urlViaCEP = servidorViaCEP.URL + "/%s"
	urlWeather = servidorWeather.URL
	defer func() {
		urlViaCEP = urlViaCEPOriginal
		urlWeather = urlWeatherOriginal
	}()

	req := httptest.NewRequest(http.MethodGet, "/01001000", nil)
	rec := httptest.NewRecorder()

	manipuladorClima(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d; esperado %d", rec.Code, http.StatusOK)
	}

	var resposta Temperaturas
	if err := json.NewDecoder(rec.Body).Decode(&resposta); err != nil {
		t.Fatalf("erro ao decodificar resposta: %v", err)
	}

	if !aproximado(resposta.TempC, 28.5) {
		t.Errorf("temp_C = %.2f; esperado 28.5", resposta.TempC)
	}
	if !aproximado(resposta.TempF, 83.3) {
		t.Errorf("temp_F = %.2f; esperado 83.3", resposta.TempF)
	}
	if !aproximado(resposta.TempK, 301.5) {
		t.Errorf("temp_K = %.2f; esperado 301.5", resposta.TempK)
	}
}
