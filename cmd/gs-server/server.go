package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
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
    if len(pArgs) != 1 {
        fmt.Println("Usage:\ngo run server.go <port>")
        return
    }
    listenPort := pArgs[0]

    startListen(listenPort)
}

func startListen(listenPort string) {
    // Listen for incoming connections
    l, err := net.Listen(CONN_TYPE, ":" + listenPort)
    if err != nil {
        fmt.Println("Error listening:", err.Error())
        os.Exit(1)
    }
    // Close the listener when the application closes
    defer l.Close()

    // Server launched
    fmt.Println("Listening on port " + listenPort)
    for {
        // Listen for an incoming connection
        conn, err := l.Accept()
        if err != nil {
            fmt.Println("Error accepting: ", err.Error())
            os.Exit(1)
        }

        // Add conn to a map containing all connected clients

        // Handle connections in a new goroutine
        go handleConnection(conn)
    }
}

func handleConnection(conn net.Conn) {
    // Loop for each new connection
    for {
        // Read data line by line
        readData, err := bufio.NewReader(conn).ReadString(EOT)
        if err != nil {
            fmt.Println("Error: ", err.Error())
            return
        }

        // Print & interpret data
        data := normalizeReadData(readData)
        fmt.Printf(">> %s\n", data)

        // Handle disconnect if a "DISCONNECT" is received
        if data == DISCONNECT {
            fmt.Println("Received a disconnect, closing connection ...")
            respondClient(conn, DISCONNECT + ":" + ACK)
            break
        }
        interpretData(conn, data)
    }
    // Close connection when out of the loop
    conn.Close()
}

func interpretData(conn net.Conn, data string) {
        if data == "ping" {
            respondClient(conn, "pong")
        } else {
            respondClient(conn, data)
        }
}

func respondClient(conn net.Conn, data string) {
    // Build the response with EOT as the last byte
    slice := normalizeWriteData(data)

    // Send response to client
    conn.Write(slice)
    fmt.Printf("<< %s\n", data)
}

func normalizeWriteData(writeData string) []byte {
    // Build the write data with EOT as the last byte

    // Init a slice with len + 1 for EOT
    slice := make([]byte, 0, len(writeData) + 1)

    // Append writeData into slice
    slice = append(slice, writeData...)

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
