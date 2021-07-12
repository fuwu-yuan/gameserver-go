package netutils

import (
	"net"

	"github.com/fuwu-yuan/gameserver-go/src/netfmt"
)

const EOT = 4

// Close the connection
func SendEotPacket(conn net.Conn) {
	eotPacket := string(append(make([]byte, 0, 1), EOT))
	conn.Write(netfmt.Output(eotPacket))
}
