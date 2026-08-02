package main

import (
	"log"

	"github.com/conceicaoG/fintrack-telegram-bot/internal/fintrackclient"
	"github.com/conceicaoG/fintrack-telegram-bot/internal/inteligencia"
	"github.com/conceicaoG/fintrack-telegram-bot/internal/telegram"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("erro ao carregar o arquivo .env:", err)
	}

	// Client da IA
	inteligenciaClient, err := inteligencia.NovoClient()
	if err != nil {
		log.Fatal(err)
	}

	// Service da IA
	inteligenciaService := inteligencia.NovoService(
		inteligenciaClient,
	)

	// Client do BFA
	fintrackClient := fintrackclient.NovoClient()

	// Inicia o Telegram
	telegram.IniciarBot(
		inteligenciaService,
		fintrackClient,
	)
}
