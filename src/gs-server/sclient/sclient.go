package sclient

import (
	"bufio"
	"fmt"
	"net"

	"github.com/fuwu-yuan/gameserver-go/src/gs-server/config"
	"github.com/fuwu-yuan/gameserver-go/src/netfmt"
)

const (
	EXT = 3
	EOT = 4
)

type Sclient struct {
	Id     string
	Socket net.Conn
}

func readLoop(reader *bufio.Reader, rChan chan string, sSettings *config.ServerSettings) {
	for sSettings.SignalToStop == false {
		// Read data until ETX (End of text)
		readData, err := reader.ReadString(EXT)
		if err != nil {
			fmt.Println("Error on read: ", err.Error())
			return
		}

		data := netfmt.Input(readData)
		rChan <- data
	}
}

// Loops for each new connection
func HandleConnection(client Sclient, sSettings *config.ServerSettings) {
	var nCurrentPlayers *uint32 = &sSettings.NCurrentPlayers
	fmt.Printf("Client (%d) connected %s\n", *nCurrentPlayers, client.Socket.RemoteAddr())

	reader := bufio.NewReader(client.Socket)
	rChan := make(chan string, 1)
	defer close(rChan)

	// Read loop
	go readLoop(reader, rChan, sSettings)

	// Interpret data and write loop
	for sSettings.SignalToStop == false {
		if len(rChan) > 0 {
			data := <-rChan

			// Handle disconnect if the first byte of a packet is EOT
			if len(data) > 0 && []byte(data)[0] == EOT {
				fmt.Println("Received a disconnect, closing connection ...") // DEBUG
				break
			} else {
				// Print & interpret data
				fmt.Printf(">> %s\n", data) // DEBUG
				interpretData(client.Socket, data)
			}
		}
	}
	// Close connection when out of the read loop
	client.Socket.Close()
	fmt.Printf("Client (%d) disconnected %s\n", *nCurrentPlayers, client.Socket.RemoteAddr())
	// Decrease number of connected clients
	*nCurrentPlayers--
}

func interpretData(conn net.Conn, data string) {
	// TODO implement protocol
	if data == "ping" {
		sendToClient(conn, "pong")
	} else {
		sendToClient(conn, data)
	}
}

func sendToClient(conn net.Conn, rawData string) {
	// Build the response with EXT as the last byte
	data := netfmt.Output(rawData)

	// Send response to client
	conn.Write(data)
	fmt.Printf("<< %s\n", rawData) // DEBUG
}
