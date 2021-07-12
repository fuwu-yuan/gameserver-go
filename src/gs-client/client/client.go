package main

import (
	"bufio"
	"fmt"
	"net"
	"os"

	"github.com/fuwu-yuan/gameserver-go/src/netfmt"
)

const (
	CONN_TYPE = "tcp"
	EXT       = 3
	EOT       = 4
)

const (
	ACK        = "ACK"
	NACK       = "NACK"
	DISCONNECT = "DISCONNECT"
)

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

	// Creating reader to read user input
	reader := bufio.NewReader(os.Stdin)
	for {
		rawLine, _ := reader.ReadString('\n')
		writeData := netfmt.Output(rawLine)
		_, err = conn.Write(writeData)
		if err != nil {
			fmt.Println("Write to server failed:", err.Error())
			os.Exit(1)
		}

		// Read server response until ETX (End of text)
		readData, err := bufio.NewReader(conn).ReadString(EXT)
		if err != nil {
			fmt.Println("Error read: ", err.Error())
			return
		}

		reply := netfmt.Input(readData)

		// Handle disconnect if the first byte of a packet is EOT
		if len(reply) > 0 && []byte(reply)[0] == EOT {
			sendEotPacket(conn)
			fmt.Println("Received a disconnect, closing connection ...") // DEBUG
			break
		} else {
			// Print & interpret data
			fmt.Printf(">> %s\n", reply)
		}
	}
	conn.Close()
	fmt.Printf("Connection to %s:%s closed", serverAddr, serverPort)
}

func sendEotPacket(conn net.Conn) {
	eotPacket := string(append(make([]byte, 0, 1), EOT))
	conn.Write(netfmt.Output(eotPacket))
}
