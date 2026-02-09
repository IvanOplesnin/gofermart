package main

import (
	"log"
	"net/http"

	"github.com/IvanOplesnin/gofermart.git/internal/config"
	"github.com/IvanOplesnin/gofermart.git/internal/handler"
	"github.com/IvanOplesnin/gofermart.git/internal/logger"
	"github.com/IvanOplesnin/gofermart.git/internal/repository/psql"
	"github.com/IvanOplesnin/gofermart.git/internal/service/gophermart"
	"github.com/IvanOplesnin/gofermart.git/internal/service/hasher"
)

func main() {
	cfg := config.InitConfig()
	log.Println(cfg)
	if err := logger.SetupLogger(&cfg.Logger); err != nil {
		log.Fatalf("error setupLogger: %s", err.Error())
	}
	
	db, err := psql.Connect(cfg.Dsn)
	if err != nil {
		logger.Log.Fatalf("db connect error: %s", err.Error())
		return
	}
	repo := psql.NewRepo(db)
	hasher := hasher.NewSHA256()

	svc := gophermart.New(cfg, hasher, repo)
	if svc == nil {
		logger.Log.Fatal("service init error")
	}

	mux := handler.InitHandler(svc)
	logger.Log.Infof("Listen on %s", cfg.RunAddress)

	if err := http.ListenAndServe(cfg.RunAddress, mux); err != nil {
		logger.Log.Fatalf("error ListenAndServe: %s", err.Error())
	}
}

