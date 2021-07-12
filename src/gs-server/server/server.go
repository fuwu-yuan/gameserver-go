package server

import (
	"fmt"
	"net"
	"os"

	"github.com/fuwu-yuan/gameserver-go/src/gs-server/config"
	"github.com/fuwu-yuan/gameserver-go/src/gs-server/sclient"
	"github.com/fuwu-yuan/gameserver-go/src/netfmt"
)

const (
	CONN_TYPE = "tcp"
	EXT       = 3
	EOT       = 4
)

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

		// Add conn to a map containing all connected clients
		client := sclient.Sclient{fmt.Sprint(settings.NCurrentPlayers), conn}
		clients[settings.NCurrentPlayers] = client

		// Increase number of connected clients
		settings.NCurrentPlayers++
		// Handle connections in a new goroutine
		go sclient.HandleConnection(client, settings) // TODO sclient.Sclient.Run()
	}

	// Disconnect all clients
	disconnectAllClients(clients)
}

func disconnectAllClients(clients map[uint32]sclient.Sclient) {
	var nbClient uint32 = uint32(len(clients))
	for i := nbClient; i > 0; i-- {
		client := clients[i]
		// If this is the "disconnect" connection from self, do not write to its socket
		// because the connection is already closed
		if i != nbClient {
			// Write EOT to all clients to end the read loop which contains net.Conn.Close()
			// This will trigger the end of the read loop
			SendEotPacket(client.Socket)
		}
		// Connection is closed in the read loop
	}
}

// Close the connection
func SendEotPacket(conn net.Conn) {
	eotPacket := string(append(make([]byte, 0, 1), EOT))
	conn.Write(netfmt.Output(eotPacket))
}
