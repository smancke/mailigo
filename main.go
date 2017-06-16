package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/smancke/mailigo/api"
	"github.com/smancke/mailigo/logging"
	"github.com/smancke/mailigo/mail"
)

const applicationName = "mailigo"

func main() {
	config := ReadConfig()
	if err := logging.Set(config.LogLevel, config.TextLogging); err != nil {
		exit(nil, err)
	}

	configForLogging := *config
	configForLogging.SMTPConfig.Password = "****"
	logging.LifecycleStart(applicationName, configForLogging)

	handlerChain := logging.NewLogMiddleware(createHandler(config))

	stop := make(chan os.Signal)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	addr := fmt.Sprintf("%v:%v", config.Host, config.Port)
	httpSrv := &http.Server{Addr: addr, Handler: handlerChain}

	go func() {
		if err := httpSrv.ListenAndServe(); err != nil {
			if err == http.ErrServerClosed {
				logging.ServerClosed(applicationName)
			} else {
				exit(nil, err)
			}
		}
	}()
	logging.LifecycleStop(applicationName, <-stop, nil)

	ctx, ctxCancel := context.WithTimeout(context.Background(), config.GracePeriod)

	httpSrv.Shutdown(ctx)
	ctxCancel()
}

func createHandler(config *Config) http.Handler {
	sender := mail.NewSMTPSender(config.SMTPConfig)
	manager := mail.NewMailingManager("templates", sender)
	h := api.NewHandler(manager)
	return h
}

var exit = func(signal os.Signal, err error) {
	logging.LifecycleStop(applicationName, signal, err)
	if err == nil {
		os.Exit(0)
	} else {
		os.Exit(1)
	}
}
