package main

import (
	"fmt"
	netHttp "net/http"
	"os"

	"primos/api"
	"primos/config"
	"primos/core"
	"primos/services"
	"primos/ui"
	"primos/ui/views/dashboard"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Carga de configuración, setup inicial
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal configuration error: %v\n", err)
		os.Exit(1)
	}

	// Dominio / Servicios
	lockService := services.NewRoomsLockService()
	updaterService := services.NewUpdaterService(cfg.GithubUser, cfg.GithubRepo)

	// Bubble Tea TUI
	initialModel := dashboard.NewModel(ui.DefaultTheme, lockService)
	program := tea.NewProgram(
		initialModel,
		tea.WithAltScreen(),
	)

	// Enrutador centralizado de API
	apiRouter := api.BuildDefaultRouter(api.AppServices{
		RoomsLock: lockService,
		Updater:   updaterService,
		OnRestart: func() {
			if program != nil {
				program.Quit()
			} else {
				os.Exit(0)
			}
		},
	})

	// Broker MQTT
	mqttBroker := core.NewMQTTBroker(cfg.MQTTAddr)
	apiRouter.RegisterMQTTRoutes(mqttBroker)

	go func() {
		if err := mqttBroker.Start(); err != nil {
			fmt.Fprintf(os.Stderr, "MQTT Server error: %v\n", err)
		}
	}()
	defer mqttBroker.Stop()

	// Servidor HTTP
	httpMux := netHttp.NewServeMux()
	apiRouter.RegisterHTTPRoutes(httpMux)

	httpServer := &netHttp.Server{
		Addr:    cfg.HTTPAddr,
		Handler: httpMux,
	}

	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != netHttp.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "HTTP Server error: %v\n", err)
		}
	}()

	if _, err := program.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Fatal runtime error: %v\n", err)
		os.Exit(1)
	}
}
