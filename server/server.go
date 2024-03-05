package server

import (
	"github.com/go-chi/chi"
	"github.com/spf13/viper"
	"log"
	"net/http"
	"scalingo_assesment/handler"
)

func SetupHTTPServer(cfg *viper.Viper) error {
	Handler, err := handler.NewHandler(cfg)
	if err != nil {
		return err
	}
	router := chi.NewRouter()
	router.Post("/login", Handler.HandleLogin)
	router.Get("/repositories", Handler.RepositoriesHandler)
	listenAddress := cfg.GetString("server.listen")
	log.Printf("Server listening on: %s\n", listenAddress)
	err = http.ListenAndServe(listenAddress, router)
	if err != nil {
		log.Fatalf("Failed to start TLS server: %v", err)
	}
	return nil
}
