package server

import (
	"fmt"
	"net"
	"time"

	"github.com/fuwu-yuan/gameserver-go/infrastructure/config"
	"github.com/fuwu-yuan/gameserver-go/netfmt"
)

const CONN_TYPE = "tcp"

type Server struct {
	port          uint16
	listener      net.Listener
	clients       map[string]*Client
	broadcastChan chan Message

	Name            string
	Description     string
	GameName        string
	GameVersion     string
	NMaxPlayers     uint
	NCurrentPlayers uint
}

type Message struct {
	SenderID string
	Message  string
}

func NewServer(settings *config.ServerSettings) (*Server, error) {
	listener, err := net.Listen(CONN_TYPE, fmt.Sprintf(":%d", settings.ServerPort))
	if err != nil {
		return nil, err
	}

	return &Server{
		port:            settings.ServerPort,
		listener:        listener,
		clients:         make(map[string]*Client),
		broadcastChan:   make(chan Message),
		Name:            settings.ServerName,
		Description:     settings.ServerDescription,
		GameName:        settings.GameName,
		GameVersion:     settings.GameVersion,
		NMaxPlayers:     settings.NMaxPlayers,
		NCurrentPlayers: settings.NCurrentPlayers,
	}, nil
}

func (s *Server) Start() error {
	fmt.Printf("Listening on port: %d\n", s.port)
	go s.broadcast()

	for {
		// Listen for an incoming connection
		conn, err := s.listener.Accept()
		if err != nil {
			if _, ok := err.(*net.OpError); ok {
				/*
				 * If the server is stoping, this gives the time for all clients to
				 * disconnect properly
				 */
				time.Sleep(time.Second * 3)
				return nil
			}
			return err
		}

		// Increase number of connected clients
		s.NCurrentPlayers++

		// Add conn to a map containing all connected clients
		client := NewClient(conn, s.NCurrentPlayers, s.broadcastChan)
		s.clients[client.ID] = client
		go func(c *Client) {
			client.Start()
			s.RemoveClient(client)
		}(client)
	}
}

func (s *Server) RemoveClient(client *Client) {
	s.NCurrentPlayers--
	delete(s.clients, client.ID)
	fmt.Printf("[%s] (%s) removing client...\n", client.RemoteAddr, client.ID) // DEBUG
}

func (s *Server) Stop() {
	for _, client := range s.clients {
		client.Close()
	}
	close(s.broadcastChan)
	s.listener.Close()
}

func (s Server) broadcast() {
	for {
		msg, ok := <-s.broadcastChan

		// If the broadcast channel has been closed there is no need to continue to loop
		if !ok {
			return
		}
		for _, client := range s.clients {
			fmt.Println("BROADCAST CLIENTS FOR START")
			// Build the response with EXT as the last byte and send it to the client
			client.Socket.Write(netfmt.Output(msg.Message))
			fmt.Println([]byte(msg.Message))
			fmt.Printf("[%s] (%s) << %s\n", client.RemoteAddr, client.ID, msg.Message) // DEBUG
		}
	}
}
