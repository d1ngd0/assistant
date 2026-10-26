package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"

	"github.com/d1ngd0/assistant/ui"
	"github.com/spf13/viper"
)

var (
	// APIRouter routes all the API endpoints for interacting with the service. This
	// includes sending context prompts etc to the assistent
	APIRouter http.ServeMux

	// UIRouter is used to route the ui requests.
	UIRouter http.ServeMux

	// Cfg holds the configuration for the application as a whole
	Cfg *viper.Viper

	// Log and ELog are used to output logs for the application
	Log  = log.New(os.Stdout, "", log.Ldate|log.Lshortfile)
	ELog = log.New(os.Stderr, "Error: ", log.Ldate|log.Lshortfile)
)

func main() {
	ParseFlags()
	Cfg = must(LoadConfigFile(_ConfigFile))

	// Setup the logging for the ui
	ui.Log = Log
	ui.ELog = ELog

	// TODO: This is fucking stupid, and a limitation of viper. We need to do
	// configuration with a different library, but that can happen another day.
	modelCfg := Cfg.Sub("models")
	for _, key := range modelCfg.AllKeys() {
		key = strings.SplitN(key, ".", 2)[0]
		if ui.Models.Has(key) {
			continue
		}

		Log.Printf("loading model %s", key)

		// Load the model in
		ui.Models.Load(key, modelCfg.Sub(key))
	}

	var rootRoute http.ServeMux
	rootRoute.Handle("/api/v1", &APIRouter)
	rootRoute.Handle("/", ui.LoadRoutes(&UIRouter))

	var srv = http.Server{
		Addr:     Cfg.GetString("http.addr"),
		Handler:  &rootRoute,
		ErrorLog: ELog,
	}

	var catch = make(chan os.Signal, 1)
	signal.Notify(catch, os.Interrupt)

	// Watch for Interupt and gracefully shutdown
	go func() {
		<-catch
		ctx, cnl := context.WithCancel(context.Background())
		go func(ctx context.Context) {
			if err := srv.Shutdown(ctx); err != nil {
				ELog.Printf("Could not shutdown server gracefully: %s", err)
			}
		}(ctx)
		<-catch
		cnl()
	}()

	Log.Printf("Running server on %s", Cfg.GetString("http.addr"))

	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		ELog.Printf("HTTP Server did not shut down gracefully: %s", err)
	}
}

// must is to be used on startup to configure global main variables.
func must[T any](v T, err error) T {
	if err != nil {
		Log.Fatalf("Could not initialize assistant: %s", err)
	}

	return v
}
