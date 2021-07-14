package server

import (
	"bufio"
	"fmt"
	"net"

	"github.com/fuwu-yuan/gameserver-go/netfmt"
	"github.com/fuwu-yuan/gameserver-go/netutils"
	uuid "github.com/satori/go.uuid"
)

const (
	EXT = 3
	EOT = 4
)

type Client struct {
	broadcastChan chan Message
	closeChan     chan bool

	ID           string
	Socket       net.Conn
	RemoteAddr   net.Addr
	IsConnected  bool
	PlayerNumber uint
}

func NewClient(conn net.Conn, playerNumber uint, broadcastChan chan Message) *Client {
	return &Client{
		broadcastChan: broadcastChan,
		closeChan:     make(chan bool),
		ID:            uuid.NewV4().String(),
		Socket:        conn,
		RemoteAddr:    conn.RemoteAddr(),
		IsConnected:   true,
		PlayerNumber:  playerNumber,
	}
}

func (c *Client) Start() {
	fmt.Printf("[%s] (%s) connected\n", c.RemoteAddr, c.ID)
	rChan := make(chan string, 1)
	defer close(rChan)

	// Read loop
	go c.readLoop(rChan)

	exit := false
	for {
		select {
		case <-c.closeChan:
			exit = true
			break
		case data := <-rChan:
			// Handle disconnect if the first byte of a packet is EOT
			if len(data) > 0 && []byte(data)[0] == EOT {
				fmt.Println("Received a disconnect, closing connection ...") // DEBUG
				exit = true
				break
			}
			// Print & interpret data
			fmt.Printf("[%s] (%s) >> %s\n", c.RemoteAddr, c.ID, data) // DEBUG
			c.broadcastChan <- Message{
				SenderID: c.ID,
				Message:  data,
			}
		}
		if exit {
			break
		}
	}

	// If out of the loop, acknowledge disconnection by sending EOT
	if err := netutils.SendEotPacket(c.Socket); err != nil {
		fmt.Printf("[%s] (%s) err: %s\n", c.RemoteAddr, c.ID, err.Error())
	}
	// Then close connection
	if err := c.Socket.Close(); err != nil {
		fmt.Printf("[%s] (%s) err: %s\n", c.RemoteAddr, c.ID, err.Error())
	}
	fmt.Printf("[%s] (%s) disconnected\n", c.RemoteAddr, c.ID)
}

func (c *Client) Close() {
	fmt.Printf("[%s] (%s) closing client...\n", c.RemoteAddr, c.ID) // DEBUG
	c.closeChan <- true
}

func (c Client) readLoop(rChan chan string) {
	reader := bufio.NewReader(c.Socket)

	select {
	case <-c.closeChan:
		break
	default:
		readData, err := reader.ReadString(EXT)
		if err != nil {
			break
		}

		data := netfmt.Input(readData)
		rChan <- data
	}
}
