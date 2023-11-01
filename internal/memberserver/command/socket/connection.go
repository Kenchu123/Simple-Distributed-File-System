package socket

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
)

var chunkSize int = 4096

func Send(conn net.Conn, msg []byte) (int, error) {
	var sent int
	for {
		if len(msg) == 0 {
			break
		}
		chunk := []byte(msg)
		if len(chunk) > chunkSize {
			chunk = chunk[:chunkSize]
		}
		n, err := conn.Write(chunk)
		if err != nil {
			return sent, fmt.Errorf("failed to write to connection: %w", err)
		}
		sent += n
		msg = msg[n:]
	}
	return sent, nil
}

func Receive(conn net.Conn) (int, []byte, error) {
	var received int
	buffer := bytes.NewBuffer(nil)
	for {
		chunk := make([]byte, chunkSize)
		read, err := conn.Read(chunk)
		if err != nil && !errors.Is(err, io.EOF) {
			return received, buffer.Bytes(), fmt.Errorf("failed to read from connection: %w", err)
		}
		received += read
		buffer.Write(chunk[:read])

		if read == 0 || read < chunkSize {
			break
		}
	}
	return received, buffer.Bytes(), nil
}
