package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"avito-shop/internal/api"
	"avito-shop/internal/config"
	"avito-shop/internal/repository"
	"avito-shop/internal/repository/db"
	"avito-shop/internal/repository/postgres"
	"avito-shop/internal/service"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	database, err := db.NewConnection(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	repos := &repository.Repositories{
		Users:        postgres.NewUserRepository(database),
		Merchandise:  postgres.NewMerchandiseRepository(database),
		Transactions: postgres.NewTransactionRepository(database),
		Inventory:    postgres.NewUserInventoryRepository(database),
	}

	services := service.NewServices(service.ServicesDeps{
		Repos:       repos,
		TokenSecret: cfg.JWT.SecretKey,
	})

	router := api.NewRouter(services)
	handler := router.Setup()

	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	log.Printf("Server starting on %s", addr)
	srv := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
