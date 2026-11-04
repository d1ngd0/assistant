package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"

	"github.com/d1ngd0/assistant/api"
	"github.com/d1ngd0/assistant/ollama"
	"github.com/d1ngd0/assistant/state"
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

	// Models defines all the models we have configured ourselves to use
	Models state.ModelRegistery

	// The active sessions that are going on.
	Sessions state.Sessions
)

func main() {
	ParseFlags()
	Cfg = must(LoadConfigFile(_ConfigFile))

	Models.Register(state.ModelTypeNoOp, func(config *viper.Viper) (state.Model, error) {
		return state.ModelFn(state.NoOpModel), nil
	})

	Models.Register(ollama.ModelType, func(config *viper.Viper) (state.Model, error) {
		var oconf ollama.Config
		if err := config.Unmarshal(&oconf); err != nil {
			return nil, fmt.Errorf(`type: "ollama": %w`, err)
		}

		return ollama.NewOllamaModel(nil, oconf)
	})

	// Setup the logging for the ui
	ui.Log = Log
	ui.ELog = ELog

	api.Log = Log
	api.ELog = ELog

	// TODO: This is fucking stupid, and a limitation of viper. We need to do
	// configuration with a different library, but that can happen another day.
	modelCfg := Cfg.Sub("models")
	for _, key := range modelCfg.AllKeys() {
		key = strings.SplitN(key, ".", 2)[0]
		if Models.Has(key) {
			continue
		}

		conf := modelCfg.Sub(key)
		Log.Printf("loading model %s", key)

		// Load the model in
		if err := Models.Load(key, conf); err != nil {
			Log.Fatalf("could not load model %s: %s", key, err)
		}
	}

	var rootRoute http.ServeMux
	rootRoute.Handle("/api/v1/", http.StripPrefix("/api/v1", api.LoadRoutes(&APIRouter, &Models, &Sessions)))
	rootRoute.Handle("/", ui.LoadRoutes(&UIRouter, &Models, &Sessions))

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
