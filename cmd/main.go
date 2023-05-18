package main

import (
	"log"
	"os"
	"./api"
)

func main() {
	logFile, err := os.OpenFile("/var/log/dockerIPAM.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer logFile.Close()

	logger := log.New(logFile, "", log.LstdFlags)

	err = api.StartAPI(logger)
	if err != nil {
		logger.Printf("Failed to start the API: %v", err)
	}
}

