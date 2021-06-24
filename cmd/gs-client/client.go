package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

const (
    CONN_TYPE = "tcp"
    EOT = 4
)

const (
    ACK = "ACK"
    NACK = "NACK"
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
    tcpAddr, err := net.ResolveTCPAddr(CONN_TYPE, serverAddr + ":" + serverPort)
    if err != nil {
        println("ResolveTCPAddr failed:", err.Error())
        os.Exit(1)
    }
    // Initiating connection
    conn, err := net.DialTCP(CONN_TYPE, nil, tcpAddr)
    if err != nil {
        println("Dial failed:", err.Error())
        os.Exit(1)
    }

    // Creating reader to read user input
    reader := bufio.NewReader(os.Stdin)
    for {
        rawLine, _ := reader.ReadString('\n')
        writeData := normalizeWriteData(rawLine)
        _, err = conn.Write(writeData)
        if err != nil {
            println("Write to server failed:", err.Error())
            os.Exit(1)
        }

        // Read server response
        readData, err := bufio.NewReader(conn).ReadString(EOT)
        if err != nil {
            fmt.Println("Error: ", err.Error())
            return
        }

        // Print message received
        reply := normalizeReadData(readData)
        fmt.Printf(">> %s\n", reply)

        // Handle disconnect if a "DISCONNECT" is received
        if reply == DISCONNECT + ":" + ACK {
            fmt.Println("Received disconnect acknowledge, closing connection ...")
            break
        }

        // Interpret data
    }
    conn.Close()
    fmt.Printf("Connection to %s:%s closed", serverAddr, serverPort)
}

func normalizeWriteData(writeData string) []byte {
    // Build the line with EOT as the last byte

    // Removes '\n'
    res := strings.ReplaceAll(writeData, "\n", "")

    // Init a slice with len + 1 for EOT
    slice := make([]byte, 0, len(res) + 1)

    // Append line into slice
    slice = append(slice, res...)

    // Add EOT at the end of slice
    slice = append(slice, EOT)

    return slice
}

func normalizeReadData(readData string) string {
    // Build the read data without EOT as the last byte

    // Get the lenght of the read data
    lenData := len(readData)

    // Init a slice with juste enough space for the actual data whithout any '\n' or EOT
    slice := make([]byte, 0, lenData)

    // Append the data with EOT
    slice = append(slice, readData...)

    // Remove the last byte (EOT)
    slice = append(slice[:lenData - 1], slice[lenData:]...)

    return string(slice)
}
