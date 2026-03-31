package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/shaokawa/merutomo/backend/internal/router"

)

func main() {
	r := gin.Default()
	router.Setup(r)

	log.Println("server started on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
