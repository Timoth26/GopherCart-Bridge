package main

import (
	"fmt"
	"log"

	"github.com/Timoth26/GopherCart-Bridge/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("failed to load config: ", err)
	}

	fmt.Printf("Starting supplier-bridge on port %s\n", cfg.Port)
	fmt.Printf("DB: %s\n", cfg.DBHost)
	fmt.Printf("Redis: %s\n", cfg.RedisAddr)
}
