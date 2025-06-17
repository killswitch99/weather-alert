package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"workflow-code-test/api/internal/execution"
	"workflow-code-test/api/internal/service"
	"workflow-code-test/api/pkg/db"
	"workflow-code-test/api/pkg/log"
	"workflow-code-test/api/pkg/models"
	"workflow-code-test/api/pkg/node"
	"workflow-code-test/api/pkg/node/condition"
	"workflow-code-test/api/pkg/node/email"
	"workflow-code-test/api/pkg/node/end"
	"workflow-code-test/api/pkg/node/form"
	"workflow-code-test/api/pkg/node/integration"
	"workflow-code-test/api/pkg/node/start"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Register all node types
func registerNodeTypes(registry *node.Registry) {
    registry.Register(models.NodeTypeStart, start.NewNode)
    registry.Register(models.NodeTypeForm, form.NewNode)
    registry.Register(models.NodeTypeIntegration, integration.NewNode)
    registry.Register(models.NodeTypeCondition, condition.NewNode)
    registry.Register(models.NodeTypeEmail, email.NewNode)
    registry.Register(models.NodeTypeEnd, end.NewNode)
    // New node types can be easily added here
}

func setupAPI(apiRouter *mux.Router, dbPool *pgxpool.Pool, engine *execution.Engine) {
	svc, err := service.NewService(dbPool, engine)
	if err != nil {
		slog.Error("Failed to create service", "error", err)
		return
	}
	svc.LoadRoutes(apiRouter, false) // isProduction=false
}

func main() {
	// Initialize the default logger
	log.InitializeLogger()
	// Connect to database using pgx
	dbURL := os.Getenv("DATABASE_URL")
	dbConfig := db.DefaultConfig()
	dbConfig.URI = dbURL
	
	if err := db.Connect(dbConfig); err != nil {
		slog.Error("Failed to connect to database", "error", err)
		return
	}
	defer db.Disconnect()
	dbPool := db.GetPool()
	nodeRegistry := node.NewRegistry()
	registerNodeTypes(nodeRegistry)
	engine := execution.NewEngine(nodeRegistry)
	// Setup router
	mainRouter := mux.NewRouter()
	apiRouter := mainRouter.PathPrefix("/api/v1").Subrouter()
	setupAPI(apiRouter, dbPool, engine)
	// Configure CORS
	corsHandler := handlers.CORS(
		handlers.AllowedOrigins([]string{"http://localhost:3003"}), // Frontend URL
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
		handlers.AllowCredentials(),
	)(mainRouter)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: corsHandler,
	}
	// Channel to listen for errors coming from the server
	serverErrors := make(chan error, 1)
	// Start the server in a goroutine
	go func() {
		slog.Info("Starting server on :8080")
		serverErrors <- srv.ListenAndServe()
	}()
	// Channel to listen for an interrupt or terminate signal from the OS
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)
	// Blocking select waiting for either a signal or an error
	select {
	case err := <-serverErrors:
		slog.Error("Server error", "error", err)

	case sig := <-shutdown:
		slog.Info("Shutdown signal received", "signal", sig)
		// Give outstanding requests 5 seconds to complete
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			slog.Error("Could not stop server gracefully", "error", err)
			srv.Close()
		}
	}
}
