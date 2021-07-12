package server

import (
	"fmt"
	"net"
	"os"

	"github.com/fuwu-yuan/gameserver-go/src/gs-server/config"
	"github.com/fuwu-yuan/gameserver-go/src/gs-server/sclient"
)

const CONN_TYPE = "tcp"

var clients = make(map[uint32]sclient.Sclient)

func StartListen(settings *config.ServerSettings) {
	portString := fmt.Sprint(settings.ServerPort)
	clients := make(map[uint32]sclient.Sclient)

	// Listen for incoming connections
	l, err := net.Listen(CONN_TYPE, ":"+portString)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	// Close the listener when the application closes
	defer l.Close()

	// Server launched
	fmt.Println("Listening on port " + portString)
	for settings.SignalToStop == false {
		// Listen for an incoming connection
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}

		// Increase number of connected clients
		settings.NCurrentPlayers++

		// Add conn to a map containing all connected clients
		// FIXME Change client.Id must be a unique generated ID
		client := sclient.Sclient{Id: fmt.Sprint(settings.NCurrentPlayers), Socket: conn, RemoteAddr: conn.RemoteAddr(), IsConnected: true}
		clients[settings.NCurrentPlayers] = client

		// Handle connections in a new goroutine
		go sclient.Run(client, settings) // TODO client.Run()
	}
}
