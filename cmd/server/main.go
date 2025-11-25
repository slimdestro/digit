package main

import (
	"context"
	"database/sql"
	"digit/internal/handler"
	"digit/internal/middleware"
	repository "digit/internal/repo"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println(".env file not found, relying on OS environment variables")
	}

	var tracer middleware.Tracer
	if os.Getenv("ENABLE_DATADOG") == "1" {
		tracer = middleware.NewDatadogTracer()
	} else {
		tracer = nil
	}

	dsn := os.Getenv("DB_USER") + ":" + os.Getenv("DB_PASSWORD") + "@tcp(" +
		os.Getenv("DB_HOST") + ":" + os.Getenv("DB_PORT") + ")/" +
		os.Getenv("DB_NAME") + "?parseTime=true&charset=utf8mb4"

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)

	repo := repository.NewMySQLBookRepository(db)
	h := handler.NewBookHandler(repo)

	mux := http.NewServeMux()

	mux.Handle("/v1/books", middleware.Chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.ListBooks(w, r)
		case http.MethodPost:
			h.CreateBook(w, r)
		default:
			http.NotFound(w, r)
		}
	}),
		middleware.APIKeyMiddleware("apitest"),
		middleware.RateLimitMiddleware(5, 10),
		middleware.SecurityMiddleware(),
		middleware.TelemetryMiddleware(tracer),
	))

	mux.Handle("/v1/books/", middleware.Chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.GetBook(w, r)
		case http.MethodPut:
			h.UpdateBook(w, r)
		case http.MethodDelete:
			h.DeleteBook(w, r)
		default:
			http.NotFound(w, r)
		}
	}),
		middleware.APIKeyMiddleware("apitest"),
		middleware.RateLimitMiddleware(5, 10),
		middleware.SecurityMiddleware(),
		middleware.TelemetryMiddleware(tracer),
	))

	server := &http.Server{
		Addr:         ":" + os.Getenv("APP_PORT"),
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	log.Println("Server is listening on http://localhost:" + os.Getenv("APP_PORT"))

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe(): %v", err)
		}
	}()

	<-done
	log.Println("Shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Shutdown Failed:%+v", err)
	}

	log.Println("Server exited gracefully")
}
