package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/nehachuha1/wbtech-tasks/internal/server"
	"github.com/nehachuha1/wbtech-tasks/pkg/log"
	"html/template"
	"net/http"
)

// Подгружаем переменные окружения, инициализиуем логгер, который будет дальше прокидываться
// ко всем управляющим структурам, а также билдим сервер. Запускается на порту :8080
func main() {
	if err := godotenv.Load("./cmd/wbtech/.env"); err != nil {
		panic(fmt.Sprintf("can't load .env: %v", err))
	}
	logger := log.NewLogger("./logs/logs.log")
	templates := template.Must(template.ParseGlob("./templates/*"))
	router := server.BuildNewServer(templates, logger)
	addr := ":8080"
	logger.Info("starting server...")
	http.ListenAndServe(addr, router)
}
