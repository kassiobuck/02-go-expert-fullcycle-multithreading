package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

func searchBrasilApi(ctx context.Context, chanel chan<- string, cep string) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://brasilapi.com.br/api/cep/v1/"+cep, nil)
	if err != nil {
		panic(err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	ioRes, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	chanel <- string(ioRes) + " - BrasilApi"
}

func searchViaCep(ctx context.Context, chanel chan<- string, cep string) {
	req, err := http.NewRequestWithContext(ctx, "GET", "http://viacep.com.br/ws/"+cep+"/json/", nil)
	if err != nil {
		panic(err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	ioRes, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	chanel <- string(ioRes) + " - ViaCep"
}

func main() {
	cep := os.Args[1:][0]
	chanel := make(chan string)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	go searchBrasilApi(ctx, chanel, cep)
	go searchViaCep(ctx, chanel, cep)

	select {
	case <-ctx.Done():
		log.Println("timeout")
	case v := <-chanel:
		log.Println(v)
		cancel()
	}
}
