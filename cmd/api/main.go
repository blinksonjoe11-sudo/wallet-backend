package main

import (
	"fmt"
	"log"
	"wallet-api/internal/app"
	"wallet-api/internal/config"
	"wallet-api/internal/server"
)

func main() {

	cfg, err := config.Load()

	if err != nil {
		log.Fatalf("Error occured loading config: %v", err)
	}

	a, err := app.NewApp()

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(a)

	router := server.NewRouter(a, cfg)

	add := fmt.Sprintf(":%s", cfg.PORT)

	if err := router.Run(add); err != nil {
		log.Fatalf("Server Error")
	}

}
