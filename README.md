# Clima por CEP

Serviço em Go que recebe um CEP, identifica a cidade (ViaCEP) e retorna a
temperatura atual (WeatherAPI) em Celsius, Fahrenheit e Kelvin.

## URL no Google Cloud Run

> Substitua pela URL gerada após o deploy:

```
https://SEU-SERVICO.run.app
```

Exemplo de uso:

```
GET https://SEU-SERVICO.run.app/01001000
```

## Endpoint

`GET /{cep}` — CEP com 8 dígitos numéricos.

### Respostas

| Cenário          | Status | Corpo                              |
| ---------------- | ------ | ---------------------------------- |
| Sucesso          | 200    | `{"temp_C":28.5,"temp_F":83.3,"temp_K":301.5}` |
| Formato inválido | 422    | `invalid zipcode`                  |
| CEP inexistente  | 404    | `can not find zipcode`             |

## Configuração

A aplicação precisa de uma chave da [WeatherAPI](https://www.weatherapi.com/)
na variável de ambiente `WEATHER_API_KEY`. A porta é lida de `PORT` (padrão `8080`).

## Rodar os testes

```sh
go test ./...
```

## Rodar localmente com Docker

```sh
docker build -t clima-por-cep .
docker run -p 8080:8080 -e WEATHER_API_KEY=sua_chave clima-por-cep
```

Depois é só acessar:

```sh
curl http://localhost:8080/01001000
```

## Rodar localmente sem Docker

```sh
export WEATHER_API_KEY=sua_chave
go run .
```

## Deploy no Google Cloud Run

```sh
gcloud run deploy clima-por-cep \
  --source . \
  --region us-central1 \
  --allow-unauthenticated \
  --set-env-vars WEATHER_API_KEY=sua_chave
```
