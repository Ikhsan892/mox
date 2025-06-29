package main

import (
	service "goodin"
	core "goodin/internal"
)

func main() {
	template := service.New(service.BaseConfig{
		DisableBanner: true,
	})

	template.App.OnBeforeApplicationBootstrapped().Add("anything", func(e core.BeforeApplicationBootstrapped) error {
		// you can add event on before application bootstrapped here
		return nil
	})

	template.App.OnAfterApplicationBootstrapped().Add("anything", func(e core.AfterApplicationBootstrapped) error {
		// you can add event on before application bootstrapped here
		return nil
	})

	template.Start()

	if !template.RootCmd.IsAvailableCommand() {
		template.App.Logger().Warn("Command not found")
	}

	// Check if application is shutdown
	template.ShutdownSignal()
}
