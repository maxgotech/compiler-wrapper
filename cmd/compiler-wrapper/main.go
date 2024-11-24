package main

import (
	"compiler-wrapper/internal/config"
	"compiler-wrapper/internal/http-server/handlers/compiler"
	"compiler-wrapper/internal/http-server/middleware/logger"
	"compiler-wrapper/internal/lib/logger/handlers/slogpretty"
	_ "compiler-wrapper/internal/lib/logger/sl"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/joho/godotenv"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	// load .env file

	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
		os.Exit(1)
	}

	// load config
	cfg := config.MustLoad()

	// create logger
	log := setupLogger((cfg.Env))

	log.Info(("starting url-shortener"), slog.String("env", cfg.Env))

	log.Debug("debug messages enabled")

	// create router
	router := chi.NewRouter()

	// id for each req
	router.Use(middleware.RequestID)

	// ip of user for req
	router.Use(middleware.RealIP)

	// logger for req
	router.Use(logger.New(log))

	// anti shutdown
	router.Use(middleware.Recoverer)

	// req parser
	router.Use(middleware.URLFormat)

	// prometheus for grafana
	router.Use(prometheusMiddleware)

	//TODO(Maxim): Add storage
	//TODO(Maxim): Add encryption
	//TODO(Maxim): Add auth
	//TODO(Maxim): Add /list with a list of user compiles
	router.Post("/run", compiler.New(log))

	router.Get("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		render.JSON(w, r, "You are in the right place")
	}))

	router.Handle("/metrics", promhttp.Handler())

	log.Info("starting server", slog.String("address", cfg.HTTPServer.Address))

	srv := &http.Server{
		Addr:         cfg.HTTPServer.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Error("failed to start server")
	}

	log.Error("server stopped")
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = setupPrettySlog()
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}

func setupPrettySlog() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}

var totalRequest = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_request_total",
		Help: "Number of get request",
	},
	[]string{"path"},
)

func prometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		next.ServeHTTP(w, r)

		totalRequest.With(prometheus.Labels{"path": "/products"}).Inc()

	})
}

func Init() {
	prometheus.Register(totalRequest)
}
