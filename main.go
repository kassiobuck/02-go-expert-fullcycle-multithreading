package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/valyala/fastjson"
)

const (
	CTX_TIMEOUT = 1 * time.Second
	SEARCH_CEP  = "01153000"
)

type requestFunc = func(context.Context, chan<- string, string)

func msgOutput(cep string, logradouro string, municipio string, bairro string, uf string, api string) string {
	if cep == "" {
		return ""
	}
	return fmt.Sprintf(
		"ENDERECO:\n- Cep: %s\n- Logradouro: %s\n- Municipio: %s\n- Bairro: %s\n- Estado: %s\n- Dados extraidos da API: %s",
		cep,
		logradouro,
		municipio,
		bairro,
		uf,
		api,
	)

}

func searchBrasilApi(ctx context.Context, channel chan<- string, cep string) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://brasilapi.com.br/api/cep/v1/"+cep, nil)
	if err != nil {
		panic(err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			fmt.Print("Context timeout: Tempo limite excedido")
			close(channel)
			return
		}
		panic(err)
	}

	defer res.Body.Close()
	ioRes, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	e := msgOutput(
		fastjson.GetString(ioRes, "cep"),
		fastjson.GetString(ioRes, "street"),
		fastjson.GetString(ioRes, "city"),
		fastjson.GetString(ioRes, "neighborhood"),
		fastjson.GetString(ioRes, "state"),
		"BrasilApi - brasilapi.com.br",
	)
	channel <- e
}

func searchViaCep(ctx context.Context, channel chan<- string, cep string) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://viacep.com.br/ws/"+cep+"/json/", nil)
	if err != nil {
		panic(err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			fmt.Print("Context timeout: Tempo limite excedido")
			close(channel)
			return
		}
		panic(err)
	}

	defer res.Body.Close()
	ioRes, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	e := msgOutput(
		fastjson.GetString(ioRes, "cep"),
		fastjson.GetString(ioRes, "logradouro"),
		fastjson.GetString(ioRes, "municipio"),
		fastjson.GetString(ioRes, "bairro"),
		fastjson.GetString(ioRes, "uf"),
		"ViaCep - viacep.com.br",
	)
	channel <- e
}

func main() {
	channel := make(chan string)
	ctx, cancel := context.WithTimeout(context.Background(), CTX_TIMEOUT)
	defer cancel()
	searchOptions := []requestFunc{searchBrasilApi, searchViaCep}

	for _, f := range searchOptions {
		go f(ctx, channel, SEARCH_CEP)
	}

	i := 0
	for ch := range channel {
		i++
		if ch != "" {
			fmt.Print(ch)
			close(channel)
		} else {
			if i == len(searchOptions) {
				fmt.Printf("O CEP %s não retornou nenhum resultado", SEARCH_CEP)
				close(channel)
			}
		}
	}
}
