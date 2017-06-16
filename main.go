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
)

const applicationName = "mailigo"

func main() {
	config := ReadConfig()
	if err := logging.Set(config.LogLevel, config.TextLogging); err != nil {
		exit(nil, err)
	}

	logging.LifecycleStart(applicationName, config)

	h := api.NewHandler()

	handlerChain := logging.NewLogMiddleware(h)

	stop := make(chan os.Signal)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	port := config.Port
	if port != "" {
		port = fmt.Sprintf(":%s", port)
	}

	httpSrv := &http.Server{Addr: port, Handler: handlerChain}

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

var exit = func(signal os.Signal, err error) {
	logging.LifecycleStop(applicationName, signal, err)
	if err == nil {
		os.Exit(0)
	} else {
		os.Exit(1)
	}
}
