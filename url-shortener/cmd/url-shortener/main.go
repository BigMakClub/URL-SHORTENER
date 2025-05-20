package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log/slog"
	"net/http"
	"os"
	"url-shorter/iternal/config"
	"url-shorter/iternal/http-server/handlers/redirect"
	"url-shorter/iternal/http-server/handlers/url/save"
	mwLogger "url-shorter/iternal/http-server/middleware/logger"
	"url-shorter/iternal/lib/logger/handlers/slogpretty"
	"url-shorter/iternal/lib/logger/sl"
	"url-shorter/iternal/storage/sqlite"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	// TODO: init config: cleanenv

	cfg := config.MustLoad()

	// TODO: init logger: slog

	log := setupLogger(cfg.Env)

	log.Info("starting url shortener", slog.String("env", cfg.Env))
	log.Debug("debug messages are enabled")

	// TODO: init storage: sqlite
	storage, err := sqlite.New(cfg.StoragePath)
	if err != nil {
		log.Error("error creating storage", sl.Err(err))
		os.Exit(1)
	}

	////id, err := storage.SaveURL("https://www.google.com/", "google")
	////
	////if err != nil {
	////	log.Error("failed to save url", sl.Err(err))
	////	os.Exit(1)
	////}
	//
	//log.Info("saved url", slog.Int64("id", id))

	//id, err = storage.SaveURL("https://www.google.com/", "google")
	//
	//

	_ = storage

	// TODO: init router: chi

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(mwLogger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	//router.Route("/url", func(r chi.Router) {
	//	r.Use(middleware.BasicAuth("url-shortener", map[string]string{
	//		cfg.HTTPServer.User: cfg.HTTPServer.Password,
	//	}))
	//	r.Post("/", save.New(log, storage))
	//})

	router.Get("/{alias}", redirect.New(log, storage))
	router.Post("/save", save.New(log, storage))

	log.Info("starting server", slog.String("address", cfg.Address))

	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	// TODO: run server:
	if err := srv.ListenAndServe(); err != nil {
		log.Error("error starting server")
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
