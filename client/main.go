package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/fuwu-yuan/gameserver-go/netfmt"
	"github.com/fuwu-yuan/gameserver-go/netutils"
)

const (
	CONN_TYPE = "tcp"
	EXT       = 3
	EOT       = 4
	NL        = 10
)

const (
	ACK        = "ACK"
	NACK       = "NACK"
	DISCONNECT = "DISCONNECT"
)

type client struct {
	Socket     net.Conn
	RemoteAddr net.Addr
}

func main() {
	// Checking program arguments
	pArgs := os.Args[1:]
	if len(pArgs) != 2 {
		fmt.Println("Usage:\ngo run client.go <server_addr> <port>")
		return
	}
	serverAddr := pArgs[0]
	serverPort := pArgs[1]

	// Resolving address
	fmt.Printf("Connecting to %s:%s\n", serverAddr, serverPort)
	tcpAddr, err := net.ResolveTCPAddr(CONN_TYPE, serverAddr+":"+serverPort)
	if err != nil {
		fmt.Println("ResolveTCPAddr failed:", err.Error())
		os.Exit(1)
	}
	// Initiating connection
	conn, err := net.DialTCP(CONN_TYPE, nil, tcpAddr)
	if err != nil {
		fmt.Println("Dial failed:", err.Error())
		os.Exit(1)
	}

	// TODO handle SIGINT

	client := client{
		Socket:     conn,
		RemoteAddr: conn.RemoteAddr(),
	}

	run(client)

	conn.Close()
	fmt.Printf("Connection to %s:%s closed", serverAddr, serverPort)
}

func readUserInputLoop(userInputChan chan string) {
	reader := bufio.NewReader(os.Stdin)

	for {
		// Read data until NL (New Line: \n)
		readData, _ := reader.ReadString(NL)

		data := strings.ReplaceAll(readData, "\n", "")
		userInputChan <- data
	}
}

func readServerResponseLoop(conn net.Conn, serverResponseChan chan string) {
	reader := bufio.NewReader(conn)

	for {
		// Read data until ETX (End of text)
		readData, err := reader.ReadString(EXT)
		if err != nil {
			// Error on read: EOF -> server ctrl-Ced
			break
		}

		data := netfmt.Input(readData)
		serverResponseChan <- data
	}
}

// Loops for each new connection
func run(client client) {
	userInputChan := make(chan string, 1)
	serverResponseChan := make(chan string, 1)
	defer close(userInputChan)
	defer close(serverResponseChan)

	// Read user input loop
	go readUserInputLoop(userInputChan)
	// Read server response loop
	go readServerResponseLoop(client.Socket, serverResponseChan)

	// Interpret data and write loop
	for {
		if len(userInputChan) > 0 {
			data := <-userInputChan

			// Print & interpret data
			fmt.Printf("[%s] << %s\n", client.RemoteAddr, data) // DEBUG
			byteData := []byte(netfmt.Output(data))
			client.Socket.Write(byteData)
		}
		if len(serverResponseChan) > 0 {
			data := <-serverResponseChan

			// Handle disconnect if the first byte of a packet is EOT
			if len(data) > 0 && []byte(data)[0] == EOT {
				netutils.SendEotPacket(client.Socket)
				fmt.Println("Received a disconnect, closing connection ...") // DEBUG
				break
			} else {
				// Print & interpret data
				fmt.Printf("[%s] >> %s\n", client.RemoteAddr, data) // DEBUG
			}
		}
	}

	// // If out of the loop, acknowledge disconnection by sending EOT
	// netutils.SendEotPacket(client.Socket)
}
