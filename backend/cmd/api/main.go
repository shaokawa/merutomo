package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/shaokawa/merutomo/backend/internal/auth"
	"github.com/shaokawa/merutomo/backend/internal/config"
	"github.com/shaokawa/merutomo/backend/internal/db"
	"github.com/shaokawa/merutomo/backend/internal/router"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	dbPool, err := db.NewPool(cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer dbPool.Close()

	r := gin.Default()

	supabaseClient := auth.NewSupabaseClient(cfg.SupabaseURL, cfg.SupabaseAnonKey)
	service := auth.NewService(supabaseClient)
	handler := auth.NewHandler(service)

	router.Setup(r, handler)

	log.Printf("server started on :%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}

