package igate

import (
	"bytes"
	"net"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("parser", func() {
	Context("buff", func() {
		It("should work", func() {
			b := []byte("abcdefg")
			bf := &buff{}
			bf.reset()
			bf.writeByte(b[0])
			Expect(bf.cnt).Should(Equal(1))

			for _, char := range b[1:] {
				bf.writeByte(char)
			}
			Expect(bf.cnt).Should(Equal(7))

			res := bf.readBytes()
			Expect(res).Should(Equal(b))
			Expect(bf.readIdx).Should(Equal(7))

			for _, char := range b {
				bf.writeByte(char)
			}
			res = bf.readBytes()
			Expect(res).Should(Equal(b))
			Expect(bf.readIdx).Should(Equal(14))
		})
	})
	Context("IGate", func() {
		It("should work", func() {
			pkt := []byte("user joe pass mama ver 123")
			pkt = append(pkt, RTN, NL)
			pkt = append(pkt, testPyld...)
			pkt = append(pkt, RTN, NL)

			ch := make(chan []byte, 1)
			mc := newmockConn(pkt)

			ig := NewIGate(ch)
			ig.HandleConn(mc)

			msg := <-ch
			Expect(msg).Should(Equal(testPyld))
		})
		It("heartbeat should work", func() {
			pkt := []byte("user joe pass mama ver 123")
			pkt = append(pkt, RTN, NL, HB, RTN, NL)

			ch := make(chan []byte, 1)
			mc := newmockConn(pkt)

			ig := NewIGate(ch)
			ig.HandleConn(mc)
		})
	})
})

type mockConn struct {
	buff *bytes.Buffer
}

func newmockConn(b []byte) *mockConn {
	return &mockConn{
		buff: bytes.NewBuffer(b),
	}
}

func (m *mockConn) Read(b []byte) (n int, err error) {
	return m.buff.Read(b)
}

func (m *mockConn) Write(b []byte) (n int, err error) { return 0, nil }

func (m *mockConn) Close() error { return nil }

func (m *mockConn) LocalAddr() net.Addr { return nil }

func (m *mockConn) RemoteAddr() net.Addr { return nil }

func (m *mockConn) SetDeadline(t time.Time) error { return nil }

func (m *mockConn) SetReadDeadline(t time.Time) error { return nil }

func (m *mockConn) SetWriteDeadline(t time.Time) error { return nil }
