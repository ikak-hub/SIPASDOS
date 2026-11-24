package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"sipasdos/app/repository"
	"sipasdos/app/service"
	"sipasdos/config"
	"sipasdos/database"
	"sipasdos/route"
)

func main() {
	config.LoadEnv()
	logger := config.NewLogger()

	jwtSecret := config.GetEnv("JWT_SECRET", "")
	if len(jwtSecret) < 32 {
		logger.Error("JWT_SECRET wajib diisi minimal 32 karakter (contoh: openssl rand -hex 32)")
		os.Exit(1)
	}
	tokenTTL := time.Duration(config.GetEnvInt("JWT_TTL_HOURS", 24)) * time.Hour

	pool, err := database.NewPool(context.Background())
	if err != nil {
		logger.Error("gagal terhubung ke database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer pool.Close()

	userRepository := repository.NewUserRepository(pool)
	authService := service.NewAuthService(userRepository, jwtSecret, tokenTTL)

	periodRepository := repository.NewPeriodRepository(pool)
	periodService := service.NewPeriodService(periodRepository)

	services := route.Services{
		Auth:   authService,
		Period: periodService,
	}

	app := config.NewApp(logger, pool, services, jwtSecret)

	port := config.GetEnv("APP_PORT", "3000")
	go func() {
		if err := app.Listen(":" + port); err != nil {
			logger.Error("server berhenti", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()
	logger.Info("server berjalan", slog.String("port", port))

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("sinyal berhenti diterima, menutup server")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := app.ShutdownWithContext(ctx); err != nil {
		logger.Error("gagal menutup server dengan rapi", slog.String("error", err.Error()))
	}
	logger.Info("server berhenti dengan rapi")
}
