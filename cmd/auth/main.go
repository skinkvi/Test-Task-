package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sirupsen/logrus"

	"github.com/skinkvi/TestTaskBackDev/internal/config"
	database "github.com/skinkvi/TestTaskBackDev/internal/db"
	"github.com/skinkvi/TestTaskBackDev/internal/handlers"
)

func main() {
	cfg, err := config.ParseConfig("./config/local.yaml")
	if err != nil {
		logrus.Errorf("Error parsing config: %v", err)
		return
	}

	logrus.Info("Config parsed successfully", cfg)

	_, err = database.ConnectToDb(cfg)
	if err != nil {
		logrus.Errorf("Error connecting to database: %v", err)
		return
	}

	if err := database.MigrateUser(); err != nil {
		log.Fatalf("Failed to migrate user table: %v", err)
	}
	if err := database.MigrateRefreshToken(); err != nil {
		log.Fatalf("Failed to migrate refresh token table: %v", err)
	}

	logrus.Info("Connected to database")

	r := chi.NewRouter()

	// устанавливает id для каждого входящего запроса
	r.Use(middleware.RequestID)

	// используется для получения ip отправителя запроса
	r.Use(middleware.RealIP)

	// не дает упасть серваку если в обработчике маршрута происходит паника
	r.Use(middleware.Recoverer)

	r.Post("/auth/issue/{userID}", func(w http.ResponseWriter, r *http.Request) {
		handlers.GenerateTokenPair(w, r, *cfg)
	})
	r.Post("/auth/refresh", func(w http.ResponseWriter, r *http.Request) {
		handlers.RefreshTokenPair(w, r, *cfg)
	})

	logrus.Info("Starting server on :8080")
	go func() {
		logrus.Fatal(http.ListenAndServe(":8080", r))
	}()

	// вынес это в отдельную горутину что бы не блокировать основной поток в случае ошибки
	gracefulStop := make(chan os.Signal, 1)
	signal.Notify(gracefulStop, syscall.SIGINT, syscall.SIGTERM)
	<-gracefulStop
	logrus.Info("Stopping server...")
	database.CloseDb()
	logrus.Info("Server stopped")
	os.Exit(0)
}
