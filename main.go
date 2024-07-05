package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/0x6flab/namegenerator"
	"github.com/chammond14/muzz/internal/db"
	"github.com/chammond14/muzz/internal/server"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)

func main() {
	slog.Info("Loading Config", "Function", "main")
	err := godotenv.Load()
	if err != nil {
		slog.Error("Failed to load Config, ending", "Function", "main")
		return
	}

	slog.Info("conn str", "str", os.Getenv("DB_CONN_STRING"))
	datastore, err := db.NewPostgresStore(os.Getenv("DB_CONN_STRING"))
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	slog.Info("===== testing connection =====")
	if testErr := datastore.TestConnection(ctx); testErr != nil {
		slog.Info("Testing error")
		panic(testErr)
	}

	slog.Info("Starting HTTP Server", "Function", "main")
	server := &server.Server{
		Store:     datastore,
		Validate:  validator.New(validator.WithRequiredStructEnabled()),
		Generator: namegenerator.NewGenerator(),
	}

	server.Start(os.Getenv("ADDR"))

}
