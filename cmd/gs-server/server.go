package main

import (
    "fmt"
    "net"
    "os"
    "bufio"
    "strings"
)

const (
    CONN_TYPE = "tcp"
)

func main() {
    // Checking program arguments
    pArgs := os.Args[1:]
    if len(pArgs) != 1 {
        fmt.Println("Usage:\ngo run server.go <port>")
        return
    }
    listenPort := pArgs[0]

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
        // Handle connections in a new goroutine
        go handleRequest(conn)
    }
}

func handleRequest(conn net.Conn) {
    // Loop for each new connection
    for {
        // Read data line by line
        netData, err := bufio.NewReader(conn).ReadString('\n')
        if err != nil {
            fmt.Println("Error: ", err.Error())
            return
        }

        // Print & interpret data
        data := strings.TrimSpace(string(netData))
        fmt.Printf(">> %s\n", data)
        if data == "DISCONNECT" {
            // Close connection if loop ends
            fmt.Println("Received a disconnect, closing connection ...")
            conn.Close()
            break
        }
    }
}
