package messagingx

import (
	"bytes"
	"encoding/binary"
	"io"
	"sync/atomic"
	"testing"

	amqp "github.com/rabbitmq/amqp091-go"
)

type prematureBodyConn struct {
	header      *bytes.Reader
	payloadRead atomic.Int64
}

func (c *prematureBodyConn) Read(buffer []byte) (int, error) {
	if c.header.Len() > 0 {
		return c.header.Read(buffer)
	}
	c.payloadRead.Store(int64(len(buffer)))
	return 0, io.EOF
}

func (*prematureBodyConn) Write(buffer []byte) (int, error) { return len(buffer), nil }
func (*prematureBodyConn) Close() error                     { return nil }

func TestRabbitMQRejectsOversizedFrameBeforeNegotiation(t *testing.T) {
	const declaredSize = 2 << 20
	header := make([]byte, 7)
	header[0] = 3 // AMQP body frame.
	binary.BigEndian.PutUint16(header[1:3], 1)
	binary.BigEndian.PutUint32(header[3:7], declaredSize)
	transport := &prematureBodyConn{header: bytes.NewReader(header)}

	connection, err := amqp.Open(transport, amqp.Config{FrameSize: 4096})
	if connection != nil {
		_ = connection.Close()
	}
	if err == nil {
		t.Fatal("accepted a body frame before connection negotiation")
	}
	if size := transport.payloadRead.Load(); size >= declaredSize {
		t.Fatalf("allocated an untrusted %d-byte payload before negotiation", size)
	}
}
