package handler

import (
	"github.com/spf13/viper"
)

type Handler struct {
	Config *viper.Viper
}

func NewHandler(cfg *viper.Viper) (*Handler, error) {
	return &Handler{
		Config: cfg,
	}, nil
}
