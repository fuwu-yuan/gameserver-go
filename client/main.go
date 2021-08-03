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

	client.Run()

	conn.Close()
	fmt.Printf("Connection to %s:%s closed", serverAddr, serverPort)
}

func (c *client) readUserInputLoop(userInputChan chan string) {
	reader := bufio.NewReader(os.Stdin)

	for {
		// Read data until NL (New Line: \n)
		readData, err := reader.ReadString(NL)
		if err != nil {
			fmt.Println("Err: ", err.Error())
		}

		data := strings.ReplaceAll(readData, "\n", "")
		userInputChan <- data
	}
}

func (c *client) readServerResponseLoop(serverResponseChan chan string) {
	reader := bufio.NewReader(c.Socket)
	serverResponseCloseChan := make(chan bool, 1)
	defer close(serverResponseCloseChan)

	exit := false
	for {
		select {
		case <-serverResponseCloseChan:
			exit = true
		default:
			// Read data until ETX (End of text)
			readData, err := reader.ReadString(EXT)
			if err != nil {
				// Error on read: EOF -> server ctrl-Ced
				serverResponseCloseChan <- true
				exit = true
				break
			}

			data := netfmt.Input(readData)
			serverResponseChan <- data
		}
		if exit {
			break
		}
	}
}

// Loops for each new connection
func (c *client) Run() {
	userInputChan := make(chan string, 1)

	serverResponseChan := make(chan string, 1)

	mainLoopCloseChan := make(chan bool, 1)

	defer close(userInputChan)
	defer close(serverResponseChan)
	defer close(mainLoopCloseChan)

	// Read user input loop
	go c.readUserInputLoop(userInputChan)
	// Read server response loop
	go c.readServerResponseLoop(serverResponseChan)

	exit := false
	// Interpret data and write loop
	for {
		select {
		case <-mainLoopCloseChan:
			exit = true
		case data := <-userInputChan:
			// Interpret data

			// Simulate "proper exit"
			if data == "exit" {
				netutils.SendEotPacket(c.Socket)
			} else {
				c.Socket.Write(netfmt.Output(data))

				// Print data
				fmt.Printf("[%s] << %s\n", c.RemoteAddr, data) // DEBUG
			}

		case data := <-serverResponseChan:
			// Handle disconnect if the first byte of a packet is EOT
			if len(data) > 0 && []byte(data)[0] == EOT {
				// netutils.SendEotPacket(c.Socket)
				fmt.Println("Received a disconnect, closing connection ...") // DEBUG
				mainLoopCloseChan <- true
				exit = true
			} else {
				// Print & interpret data
				fmt.Printf("[%s] >> %s\n", c.RemoteAddr, data) // DEBUG
			}
		}
		if exit {
			break
		}
	}

	// // If out of the loop, acknowledge disconnection by sending EOT
	// netutils.SendEotPacket(client.Socket)
}
