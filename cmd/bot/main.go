package main

import (
	"log"

	"github.com/conceicaoG/fintrack-telegram-bot/internal/telegram"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Erro ao carregar o arquivo .env:", err)
	}

	telegram.IniciarBot()
}
