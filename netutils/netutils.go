package netutils

import (
	"net"

	"github.com/fuwu-yuan/gameserver-go/netfmt"
)

const EOT = 4

// Close the connection
func SendEotPacket(conn net.Conn) error {
	eotPacket := string(append(make([]byte, 0, 1), EOT))
	_, err := conn.Write(netfmt.Output(eotPacket))
	if err != nil {
		return err
	}
	return nil
}
