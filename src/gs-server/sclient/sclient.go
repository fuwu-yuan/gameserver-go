package sclient

import (
	"bufio"
	"fmt"
	"net"

	"github.com/fuwu-yuan/gameserver-go/src/gs-server/config"
	"github.com/fuwu-yuan/gameserver-go/src/netfmt"
	"github.com/fuwu-yuan/gameserver-go/src/netutils"
)

const (
	EXT = 3
	EOT = 4
)

type Sclient struct {
	Id           string
	Socket       net.Conn
	RemoteAddr   net.Addr
	IsConnected  bool
	PlayerNumber uint32
}

func readLoop(client *Sclient, rChan chan string, sSettings *config.ServerSettings) {
	reader := bufio.NewReader(client.Socket)

	for sSettings.SignalToStop == false && client.IsConnected == true {
		// Read data until ETX (End of text)
		readData, err := reader.ReadString(EXT)
		if err != nil {
			// Error on read: EOF -> client ctrl-Ced
			client.IsConnected = false
			break
		}

		data := netfmt.Input(readData)
		rChan <- data
	}
}

// Loops for each new connection
func Run(client Sclient, sSettings *config.ServerSettings, clients map[string]Sclient) {
	var nCurrentPlayers *uint32 = &sSettings.NCurrentPlayers
	fmt.Printf("[%s] (%s) connected\n", client.RemoteAddr, client.Id)

	rChan := make(chan string, 1)
	defer close(rChan)

	// Read loop
	go readLoop(&client, rChan, sSettings)

	// Interpret data and write loop
	for sSettings.SignalToStop == false && client.IsConnected == true {
		if len(rChan) > 0 {
			data := <-rChan

			// Handle disconnect if the first byte of a packet is EOT
			if len(data) > 0 && []byte(data)[0] == EOT {
				fmt.Println("Received a disconnect, closing connection ...") // DEBUG
				break
			} else {
				// Print & interpret data
				fmt.Printf("[%s] (%s) >> %s\n", client.RemoteAddr, client.Id, data) // DEBUG
				interpretData(client, data, clients)
			}
		}
	}
	// If out of the loop, acknowledge disconnection by sending EOT
	netutils.SendEotPacket(client.Socket)
	// Then close connection
	client.Socket.Close()
	fmt.Printf("[%s] (%s) disconnected\n", client.RemoteAddr, client.Id)
	// Delete client from the map
	delete(clients, client.Id)
	// Decrease number of connected clients
	*nCurrentPlayers--
}

func interpretData(c Sclient, data string, clients map[string]Sclient) {
	// TODO implement protocol
	if data == "ping" {
		sendToClient(c, "pong")
	} else {
		broadcast(clients, c, data)
	}
}

func sendToClient(c Sclient, data string) {
	// Build the response with EXT as the last byte
	byteData := netfmt.Output(data)

	// Send response to client
	c.Socket.Write(byteData)
	fmt.Printf("[%s] (%s) << %s\n", c.RemoteAddr, c.Id, data) // DEBUG
}

func broadcast(clients map[string]Sclient, sender Sclient, data string) {
	for _, v := range clients {
		sendToClient(v, data)
	}
}
