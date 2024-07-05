package main

import (
	"log/slog"
	"os"
	"sso/internal/app"
	"sso/internal/config"
	"sso/internal/lib/logger/handlers/slogpretty"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	// TODO: инициализировать объект конфига
	cfg := config.MustLoad()

	//TODO: инициализировать logger
	log := setupPrettySlog(cfg.Env)
	log.Info("starting application", slog.Any("config", cfg))
	//TODO: инициализировать приложение (app)
	application := app.New(log, cfg.GRPC.Port, cfg.StoragePath, cfg.TokenTTL)
	application.GRPCSrv.MustRun()

	//TODO: запустить gRPC сервер
}

func setupPrettySlog(env string) *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}
	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}
