package signals

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fuwu-yuan/gameserver-go/src/gs-server/config"
	"github.com/fuwu-yuan/gameserver-go/src/netutils"
)

const (
	LOCALHOST = "127.0.0.1"
	CONN_TYPE = "tcp"
)

// HandleSignals will notify when either SIGINT, SIGTERM or SIGUSR1 is received by the program
func HandleSignals(sSettings *config.ServerSettings) {
	signalChannel := make(chan os.Signal, 3)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM, syscall.SIGUSR1)

	go func() {
		sig := <-signalChannel
		if sig == os.Interrupt || sig == syscall.SIGTERM || sig == syscall.SIGUSR1 {
			// Handle SIGINT | SIGTERM | SIGUSR1
			fmt.Println("Received a shutdown request, initiating shutdown sequence ...")
			sSettings.SignalToStop = true

			// Trigger a connection on self to "trigger" the accept()
			tcpAddr, _ := net.ResolveTCPAddr(CONN_TYPE, LOCALHOST+":"+fmt.Sprint(sSettings.ServerPort))
			conn, _ := net.DialTCP(CONN_TYPE, nil, tcpAddr)
			netutils.SendEotPacket(conn)
			time.Sleep(1 * time.Second)
			conn.Close()
		}
	}()
}
