package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fuwu-yuan/gameserver-go/net_io"
)

const (
    LOCALHOST = "127.0.0.1"
    CONN_TYPE = "tcp"
    EXT = 3
    EOT = 4
)

const (
    ACK = "ACK"
    NACK = "NACK"
    DISCONNECT = "DISCONNECT"
)

type serverSettings struct {
    rawArgServerPort int
    rawArgNMaxPlayers int

    serverPort uint16
    serverName string
    serverDescription string
    gameName string
    gameVersion string
    nMaxPlayers uint32
    nCurrentPlayers uint32
}

type client struct {
    id string
    socket net.Conn
}
var clients = make(map[uint32] client)

var signalToStop bool = false

func main() {
    // Settings initialization
    settings := checkArguments()

    // Signal handler initialize
    handleSignals(settings.serverPort)

    // Start the server
    startListen(&settings, clients)

    waitAllDisconnect(&settings)
}

// Wait for all disconnection of connected clients or 10 seconds and exit
func waitAllDisconnect(settings *serverSettings) {
    for i := 0; settings.nCurrentPlayers > 0 && i < 3; i++ {
        time.Sleep(1 * time.Second)
    }
    fmt.Println("Shutdown sequence complete, exiting ...")
}

// Checking program arguments
func checkArguments() serverSettings {
    var s serverSettings

    flag.IntVar(&s.rawArgServerPort, "p", -1, "port")
    flag.StringVar(&s.serverName, "sn", "", "sname")
    flag.StringVar(&s.serverDescription, "sd", "", "sdesc") // Is optional
    flag.StringVar(&s.gameName, "gn", "", "gname")
    flag.StringVar(&s.gameVersion, "gv", "", "gver")
    flag.IntVar(&s.rawArgNMaxPlayers, "nmp", -1, "nmp")
    flag.Usage = serverUsage
    flag.Parse()

    // Initialize the numbers
    if s.rawArgServerPort > 2000 && s.rawArgServerPort < 65535 {
        s.serverPort = uint16(s.rawArgServerPort)
    }
    if s.rawArgNMaxPlayers > 0 {
        s.nMaxPlayers = uint32(s.rawArgNMaxPlayers)
    }

    // Check for invalid default argument
    if s.rawArgServerPort == -1 || s.serverName == "" || s.gameName == "" ||
        s.gameVersion == "" || s.rawArgNMaxPlayers == -1 || s.serverPort == 0 ||
        s.nMaxPlayers == 0 {
        serverUsage()
        os.Exit(-1)
    }

    return s
}

func serverUsage() {
    fmt.Println("Usage:")
    fmt.Println("./gs-server <port> <server_name> <server_description> <game_name> <game_version> <max_number_of_players>")
}

func handleSignals(serverPort uint16) {
    signalChannel := make(chan os.Signal, 3)
    signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM, syscall.SIGUSR1)

    go func() {
        sig := <-signalChannel
        if sig == os.Interrupt || sig == syscall.SIGTERM || sig == syscall.SIGUSR1 {
            // Handle SIGINT | SIGTERM | SIGUSR1
            fmt.Println("Received a shutdown request, initiating shutdown sequence ...")
            signalToStop = true

            // Trigger a connection on self to "trigger" the accept()
            tcpAddr, _ := net.ResolveTCPAddr(CONN_TYPE, LOCALHOST + ":" + fmt.Sprint(serverPort))
            conn, _ := net.DialTCP(CONN_TYPE, nil, tcpAddr)
            sendEotPacket(conn)
            time.Sleep(1 * time.Second)
            conn.Close()
        }
    }()
}

func startListen(settings *serverSettings, clients map[uint32] client) {
    portString := fmt.Sprint(settings.serverPort)

    // Listen for incoming connections
    l, err := net.Listen(CONN_TYPE, ":" + portString)
    if err != nil {
        fmt.Println("Error listening:", err.Error())
        os.Exit(1)
    }
    // Close the listener when the application closes
    defer l.Close()

    // Server launched
    fmt.Println("Listening on port " + portString)
    for signalToStop == false {
        // Listen for an incoming connection
        conn, err := l.Accept()
        if err != nil {
            fmt.Println("Error accepting: ", err.Error())
            os.Exit(1)
        }

        // Add conn to a map containing all connected clients
        client := client{fmt.Sprint(settings.nCurrentPlayers), conn}
        clients[settings.nCurrentPlayers] = client

        // Increase number of connected clients
        settings.nCurrentPlayers++
        // Handle connections in a new goroutine
        go handleConnection(client, &settings.nCurrentPlayers)
    }

    // Disconnect all clients
    disconnectAllClients(clients)
}

func readLoop(reader *bufio.Reader, rChan chan string) {
    for signalToStop == false {
        // Read data until ETX (End of text)
        readData, err := reader.ReadString(EXT)
        if err != nil {
            fmt.Println("Error on read: ", err.Error())
            return
        }

        data := net_io.NormalizeReadData(readData)
        rChan <- data
    }
}

// Loops for each new connection
func handleConnection(client client, nCurrentPlayers *uint32) {
    fmt.Printf("Client (%d) connected %s\n", *nCurrentPlayers, client.socket.RemoteAddr())

    reader := bufio.NewReader(client.socket)
    rChan := make(chan string, 1)
    defer close(rChan)

    // Read loop
    go readLoop(reader, rChan)

    // Interpret data and write loop
    for signalToStop == false {
        if len(rChan) > 0 {
            data := <- rChan

            // Handle disconnect if the first byte of a packet is EOT
            if len(data) > 0 && []byte(data)[0] == EOT {
                fmt.Println("Received a disconnect, closing connection ...") // DEBUG
                break
            } else {
                // Print & interpret data
                fmt.Printf(">> %s\n", data) // DEBUG
                interpretData(client.socket, data)
            }
        }
    }
    // Close connection when out of the read loop
    client.socket.Close()
    fmt.Printf("Client (%d) disconnected %s\n", *nCurrentPlayers, client.socket.RemoteAddr())
    // Decrease number of connected clients
    *nCurrentPlayers--
}

func disconnectAllClients(clients map[uint32] client) {
    var nbClient uint32 = uint32(len(clients))
    for i := nbClient; i > 0; i-- {
        client := clients[i]
        // If this is the "disconnect" connection from self, do not write to its socket
        // because the connection is already closed
        if i != nbClient {
            // Write EOT to all clients to end the read loop which contains net.Conn.Close()
            // This will trigger the end of the read loop
            sendEotPacket(client.socket)
        }
        // Connection is closed in the read loop
    }
}

// Close the connection
func sendEotPacket(conn net.Conn) {
    eotPacket := string(append(make([]byte, 0, 1), EOT))
    conn.Write(net_io.NormalizeWriteData(eotPacket))
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
    data := net_io.NormalizeWriteData(rawData)

    // Send response to client
    conn.Write(data)
    fmt.Printf("<< %s\n", rawData) // DEBUG
}
