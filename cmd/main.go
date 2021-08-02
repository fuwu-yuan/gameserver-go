package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/fuwu-yuan/gameserver-go/infrastructure/config"
	"github.com/fuwu-yuan/gameserver-go/server"
)

func main() {
	// Settings initialization
	settings, err := config.CheckArguments()
	if err != nil {
		fmt.Printf("CheckArguments Error: %s\n", err.Error())
		os.Exit(-1)
	}
	s, err := server.NewServer(settings)
	if err != nil {
		fmt.Printf("NewServer Error: %s\n", err.Error())
		os.Exit(-1)
	}

	// Signal handler initialize
	go handleSignals(s)

	// Start the server
	if err := s.Start(); err != nil {
		fmt.Printf("Server.Start Error: %s\n", err.Error())
		os.Exit(-1)
	}
}

func handleSignals(s *server.Server) {
	signalChannel := make(chan os.Signal, 3)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM, syscall.SIGUSR1)

	<-signalChannel
	// Handle SIGINT | SIGTERM | SIGUSR1
	fmt.Println("Received a shutdown request, initiating shutdown sequence ...")
	s.Stop()
}
