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
	readChan := make(chan string, 1)
	defer close(readChan)

	// Read loop
	go c.readLoop(readChan)

	exit := false
	for {
		select {
		case <-c.closeChan:
			exit = true
		case data := <-readChan:
			// Handle disconnect if the first byte of a packet is EOT
			if len(data) > 0 && []byte(data)[0] == EOT {
				fmt.Println("Received a disconnect, closing connection ...") // DEBUG
				exit = true
			} else {
				// Print & interpret data
				fmt.Printf("[%s] (%s) >> %s\n", c.RemoteAddr, c.ID, data) // DEBUG
				c.broadcastChan <- Message{
					SenderID: c.ID,
					Message:  data,
				}
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
	c.closeChan <- true
}

func (c Client) readLoop(readChan chan string) {
	reader := bufio.NewReader(c.Socket)

	exit := false
	for {
		select {
		case <-c.closeChan:
			exit = true
		default:
			readData, err := reader.ReadString(EXT)
			if err != nil {
				exit = true
			} else {
				data := netfmt.Input(readData)
				readChan <- data
			}
		}
		if exit {
			break
		}
	}
	/*
	 * If the distant client close its socket without notice, this loop will
	 * end due to the ReadString returning an EOF error therefore we need to
	 * manually close its server-side socket
	 */
	c.Close()
}
