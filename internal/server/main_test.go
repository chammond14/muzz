package server

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/0x6flab/namegenerator"
	"github.com/chammond14/muzz/internal/db"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)

var TestServer Server

func TestMain(m *testing.M) {
	slog.Info("Loading Config", "Function", "TestMain")
	err := godotenv.Load("../../.env")
	if err != nil {
		slog.Error("Failed to load Config, ending", "Function", "main")
		return
	}

	slog.Info("Got config", "str", os.Getenv("TEST_DB_CONN_STRING"))
	datastore, err := db.NewPostgresStore(os.Getenv("TEST_DB_CONN_STRING"))
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	slog.Info("===== testing connection =====")
	if testErr := datastore.TestConnection(ctx); testErr != nil {
		slog.Info("Testing error")
		panic(testErr)
	}

	TestServer = Server{
		Store:     datastore,
		Validate:  validator.New(validator.WithRequiredStructEnabled()),
		Generator: namegenerator.NewGenerator(),
	}

	code := m.Run()
	os.Exit(code)
}
