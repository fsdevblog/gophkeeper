package main

import (
	"github.com/fsdevblog/gophkeeper/internal/app"
	"github.com/fsdevblog/gophkeeper/internal/config"
	"github.com/fsdevblog/gophkeeper/internal/logs"
	"go.uber.org/zap"
)

func main() {
	conf := config.MustLoadConfig()
	logger := logs.MustNew()

	logger.Info("Starting gophkeeper", zap.Any("config", conf))

	gophKeeper := app.Must(app.New(conf, logger))
	if err := gophKeeper.Run(); err != nil {
		logger.Error("Error running gophkeeper", zap.Error(err))
	}
}
