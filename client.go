package main

import (
    "fmt"
    "net"
    "os"
    "bufio"
//    "strings"
)

const (
    CONN_TYPE = "tcp"
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

    // Initiating connection
    fmt.Printf("Connecting to %s:%s\n", serverAddr, serverPort)
    tcpAddr, err := net.ResolveTCPAddr(CONN_TYPE, serverAddr + ":" + serverPort)
    if err != nil {
        println("ResolveTCPAddr failed:", err.Error())
        os.Exit(1)
    }
    conn, err := net.DialTCP(CONN_TYPE, nil, tcpAddr)
    if err != nil {
        println("Dial failed:", err.Error())
        os.Exit(1)
    }

    // Creating reader to read user input
    reader := bufio.NewReader(os.Stdin)
    for {
        line, _ := reader.ReadString('\n')
        _, err = conn.Write([]byte(line))
        if err != nil {
            println("Write to server failed:", err.Error())
            os.Exit(1)
        }

    // TODO
    // Read server response
/*
        // Read data line by line
        netData, err := bufio.NewReader(conn).ReadString('\n')
        if err != nil {
            fmt.Println("Error: ", err.Error())
            return
        }

        // Print message received
        data := string(netData)
        fmt.Println(">> %s", data)

        // Interpret data
        reply := strings.TrimSpace(data)
        if reply == "DISCONNECT" {
            fmt.Println("Received disconnect command, exiting ...")
            break
        }
*/
    }
    conn.Close()
    fmt.Println("Connection to %s:%s closed", serverAddr, serverPort)
}
