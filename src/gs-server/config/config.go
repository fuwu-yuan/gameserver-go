package config

import (
	"flag"
	"fmt"
	"os"
)

type ServerSettings struct {
	rawArgServerPort  int
	rawArgNMaxPlayers int

	ServerPort        uint16
	ServerName        string
	ServerDescription string
	GameName          string
	GameVersion       string
	NMaxPlayers       uint32
	NCurrentPlayers   uint32

	SignalToStop bool
}

// CheckArguments checks program arguments
func CheckArguments() ServerSettings {
	var s ServerSettings

	flag.IntVar(&s.rawArgServerPort, "p", -1, "port")
	flag.StringVar(&s.ServerName, "sn", "", "sname")
	flag.StringVar(&s.ServerDescription, "sd", "", "sdesc") // Is optional
	flag.StringVar(&s.GameName, "gn", "", "gname")
	flag.StringVar(&s.GameVersion, "gv", "", "gver")
	flag.IntVar(&s.rawArgNMaxPlayers, "nmp", -1, "nmp")
	flag.Usage = serverUsage
	flag.Parse()

	// Initialize the numbers
	if s.rawArgServerPort > 2000 && s.rawArgServerPort < 65535 {
		s.ServerPort = uint16(s.rawArgServerPort)
	}
	if s.rawArgNMaxPlayers > 0 {
		s.NMaxPlayers = uint32(s.rawArgNMaxPlayers)
	}

	// Check for invalid default argument
	if s.rawArgServerPort == -1 || s.ServerName == "" || s.GameName == "" ||
		s.GameVersion == "" || s.rawArgNMaxPlayers == -1 || s.ServerPort == 0 ||
		s.NMaxPlayers == 0 {
		serverUsage()
		os.Exit(-1)
	}
	s.SignalToStop = false

	return s
}

// serverUsage prints the usage
func serverUsage() {
	fmt.Println("Usage:")
	fmt.Println("./gs-server <port> <server_name> <server_description> <game_name> <game_version> <max_number_of_players>")
}
