package main

import (
	"log"

	"github.com/IvanOplesnin/gofermart.git/internal/config"
	"github.com/IvanOplesnin/gofermart.git/internal/logger"
)

func main() {
	cfg := config.InitConfig()
	log.Println(cfg)
	if err := logger.SetupLogger(&cfg.Logger); err != nil {
		log.Fatalf("error setupLogger: %s", err.Error())
	}
	
	// TODO: init db
	// TODO: init router
	// TODO: init server
	// TODO: run server
}

