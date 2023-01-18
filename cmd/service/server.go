package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// serve intializes the service's server and spins it up.
func (svc *service) serve() error {
	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", svc.config.Port),
		Handler:      svc.routes(),
		ErrorLog:     log.New(svc.logger, "", 0),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	shutdownErr := make(chan error)

	// Background job to listen for any shutdown signal
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit

		svc.logger.PrintInfo("shutting down server", map[string]string{
			"signal": s.String(),
		})

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := server.Shutdown(ctx)
		if err != nil {
			shutdownErr <- err
		}

		svc.logger.PrintInfo("completing background tasks", map[string]string{
			"addr": server.Addr,
		})

		svc.wg.Wait()
		shutdownErr <- nil
	}()

	svc.logger.PrintInfo("starting server", map[string]string{
		"env":  svc.config.Env,
		"addr": server.Addr,
	})

	err := server.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	err = <-shutdownErr
	if err != nil {
		return err
	}

	svc.logger.PrintInfo("server stopped", map[string]string{
		"addr": server.Addr,
	})

	return nil
}
