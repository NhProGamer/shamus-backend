package main

import (
	"log"
	"shamus-backend/internal/infrastructure/config"
)

func main() {
	var err error
	var Configuration config.Config
	Configuration, err = config.LoadConfig()
	if err != nil {
		log.Fatalf("Error while loading config file: %v", err)
	}

	if Configuration.Debug {
	} else {
	}

}
