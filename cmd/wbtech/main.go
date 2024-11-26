package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/nehachuha1/wbtech-tasks/internal/server"
	"github.com/nehachuha1/wbtech-tasks/pkg/log"
	"html/template"
	"net/http"
	"time"
)

func main() {
	if err := godotenv.Load("./cmd/wbtech/.env"); err != nil {
		panic(fmt.Sprintf("can't load .env: %v", err))
	}
	logger := log.NewLogger("logs/logs.log")
	templates := template.Must(template.ParseGlob("./templates/*"))
	router := server.BuildNewServer(templates, logger)
	addr := ":8080"
	logger.Infow("starting server...",
		"source", "cmd/wbtech/main", "time", time.Now().String())
	http.ListenAndServe(addr, router)
}
