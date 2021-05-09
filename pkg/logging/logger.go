package logging

import (
	"fmt"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
)

func GetLogger(prod bool) logr.Logger {
	var logger *zap.Logger
	var err error

	if prod {
		logger, err = zap.NewProduction()
	} else {
		logger, err = zap.NewDevelopment()
	}
	if err != nil {
		panic(fmt.Sprintf("who watches the watchmen (%v)?", err))
	}
	return zapr.NewLogger(logger)
}
