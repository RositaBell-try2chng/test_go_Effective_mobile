package main

import (
	"fmt"
	"net/http"
	"os"
	"subscription-aggregator/internal/config"
	"subscription-aggregator/internal/database"
	"subscription-aggregator/internal/handlers"
	"subscription-aggregator/internal/middleware"
	"subscription-aggregator/internal/validation"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	logger := setupLogger(cfg)

	db, err := database.New(cfg)
	if err != nil {
		logger.WithError(err).Fatal("Failed to connect to database")
	}
	defer db.Close()

	logger.Info("Successfully connected to database")

	subscriptionHandler := handlers.NewSubscriptionHandler(db, logger)

	router := mux.NewRouter()

	router.Use(middleware.LoggingMiddleware)

	// Health check
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "healthy"}`))
	}).Methods("GET")


	router.HandleFunc("/subscriptions", subscriptionHandler.ListSubscriptions).Methods("GET")

	// Create subscription
	createSubRouter := router.Path("/subscriptions").Subrouter()
	createSubRouter.Use(middleware.ValidationMiddleware(validation.GetAllowedFieldsForCreate()))
	createSubRouter.HandleFunc("", subscriptionHandler.CreateSubscription).Methods("POST")

	// Update subscription
	updateSubRouter := router.Path("/subscriptions/{id}").Subrouter()
	updateSubRouter.Use(middleware.ValidationMiddleware(validation.GetAllowedFieldsForUpdate()))
	updateSubRouter.HandleFunc("", subscriptionHandler.UpdateSubscription).Methods("PUT")

	router.HandleFunc("/subscriptions/{id}", subscriptionHandler.GetSubscription).Methods("GET")
	router.HandleFunc("/subscriptions/{id}", subscriptionHandler.DeleteSubscription).Methods("DELETE")

	// Aggregate subscriptions
	aggregateSubRouter := router.Path("/subscriptions/aggregate").Subrouter()
	aggregateSubRouter.Use(middleware.ValidationMiddleware(validation.GetAllowedFieldsForAggregation()))
	aggregateSubRouter.HandleFunc("", subscriptionHandler.AggregateSubscriptions).Methods("POST")

	// Start
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	logger.WithField("addr", addr).Info("Starting server")

	if err := http.ListenAndServe(addr, router); err != nil {
		logger.WithError(err).Fatal("Failed to start server")
	}
}

func setupLogger(cfg *config.Config) *logrus.Logger {
	logger := logrus.New()

	level, err := logrus.ParseLevel(cfg.Logging.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	if cfg.Logging.Format == "json" {
		logger.SetFormatter(&logrus.JSONFormatter{})
	} else {
		logger.SetFormatter(&logrus.TextFormatter{})
	}

	return logger
}
