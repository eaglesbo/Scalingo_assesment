package main

import (
	"scalingo_assesment/config"
	"scalingo_assesment/server"
)

func main() {
	cfg, err := config.ReadConfig()
	if err != nil {
		panic(err)
	}

	err = server.SetupHTTPServer(cfg)
	if err != nil {
		panic(err)
	}
}
