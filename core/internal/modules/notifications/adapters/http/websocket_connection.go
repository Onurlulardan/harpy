package notificationhttp

import (
	"bufio"
	"net"
	"sync"
)

const socketQueueCapacity = 32

type socketConnection struct {
	connection net.Conn
	reader     *bufio.Reader
	send       chan []byte
	done       chan struct{}
	closeOnce  sync.Once
	writeMutex sync.Mutex
}

func newSocketConnection(connection net.Conn, reader *bufio.Reader) *socketConnection {
	return &socketConnection{connection: connection, reader: reader,
		send: make(chan []byte, socketQueueCapacity), done: make(chan struct{})}
}

func (socket *socketConnection) offer(payload []byte) {
	copyPayload := append([]byte(nil), payload...)
	select {
	case socket.send <- copyPayload:
	default:
		// Realtime is a best-effort hint; canonical state remains in the inbox.
	}
}

func (socket *socketConnection) run() error {
	go socket.writeLoop()
	for {
		opcode, payload, err := readClientFrame(socket.reader)
		if err != nil || opcode == socketOpcodeClose {
			return nil
		}
		if opcode == socketOpcodePing {
			if socket.writeFrame(socketOpcodePong, payload) != nil {
				return nil
			}
		}
	}
}

func (socket *socketConnection) writeLoop() {
	for {
		select {
		case payload := <-socket.send:
			if socket.writeFrame(socketOpcodeText, payload) != nil {
				socket.close()
				return
			}
		case <-socket.done:
			return
		}
	}
}

func (socket *socketConnection) close() {
	socket.closeOnce.Do(func() {
		close(socket.done)
		_ = socket.connection.Close()
	})
}

func (socket *socketConnection) writeFrame(opcode byte, payload []byte) error {
	socket.writeMutex.Lock()
	defer socket.writeMutex.Unlock()
	return writeServerFrame(socket.connection, opcode, payload)
}
