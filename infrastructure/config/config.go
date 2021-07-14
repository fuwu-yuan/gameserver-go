package config

import (
	"errors"
	"flag"
	"fmt"
)

type ServerSettings struct {
	ServerPort        uint16
	ServerName        string
	ServerDescription string
	GameName          string
	GameVersion       string
	NMaxPlayers       uint
	NCurrentPlayers   uint

	SignalToStop bool
}

func (s ServerSettings) Validate() error {
	if s.ServerPort < 2000 && s.ServerPort > 65535 {
		return errors.New("invalid-port")
	}
	if s.ServerName == "" || s.GameName == "" || s.GameVersion == "" || s.NMaxPlayers == 0 {
		return errors.New("invalid-parameters")
	}
	return nil
}

// CheckArguments checks program arguments
func CheckArguments() (*ServerSettings, error) {
	var s ServerSettings
	var rawArgServerPort uint

	flag.UintVar(&rawArgServerPort, "p", 0, "port")
	flag.StringVar(&s.ServerName, "sn", "", "sname")
	flag.StringVar(&s.ServerDescription, "sd", "", "sdesc") // Is optional
	flag.StringVar(&s.GameName, "gn", "", "gname")
	flag.StringVar(&s.GameVersion, "gv", "", "gver")
	flag.UintVar(&s.NMaxPlayers, "nmp", 0, "nmp")
	flag.Usage = serverUsage
	flag.Parse()

	s.ServerPort = uint16(rawArgServerPort)

	// Check for invalid default argument
	if err := s.Validate(); err != nil {
		serverUsage()
		return nil, err
	}

	return &s, nil
}

// serverUsage prints the usage
func serverUsage() {
	fmt.Println("Usage:")
	fmt.Println("./gs-server <port> <server_name> <server_description> <game_name> <game_version> <max_number_of_players>")
	fmt.Println("                                 [    IS OPTIONAL   ]")
}
