package main

import (
	"fmt"
	"time"

	"github.com/fuwu-yuan/gameserver-go/src/gs-server/config"
	"github.com/fuwu-yuan/gameserver-go/src/gs-server/server"
	"github.com/fuwu-yuan/gameserver-go/src/gs-server/signals"
)

func main() {
	// Settings initialization
	settings := config.CheckArguments()

	// Signal handler initialize
	signals.HandleSignals(&settings)

	// Start the server
	server.StartListen(&settings)

	waitAllDisconnect(&settings)
}

// waitAllDisconnect waits for all disconnection of connected clients or 3 seconds and exit
func waitAllDisconnect(settings *config.ServerSettings) {
	for i := 0; settings.NCurrentPlayers > 0 && i < 3; i++ {
		time.Sleep(1 * time.Second)
	}
	fmt.Println("Shutdown sequence complete, exiting ...")
}
