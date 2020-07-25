package igate

import (
	"errors"
	"net"
	"time"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("server", func() {
	var uConn *MockConn
	var listener *MockListener
	// var tcpConn *MockConn
	BeforeEach(func() {
		ctrl := gomock.NewController(GinkgoT())
		uConn = NewMockConn(ctrl)
		listener = NewMockListener(ctrl)
		// tcpConn = NewMockConn(ctrl)
	})
	Context("NewServer", func() {
		It("should work", func() {
			s, err := NewServer("127.0.0.1:1234", "127.0.0.1:1234")
			Expect(err).Should(BeNil())

			s.tcpListener = listener
			listener.EXPECT().Close()
			s.Stop()
		})
		It("should fail for bad UDP port", func() {
			_, err := NewServer("127.0.0.1:1234", "127.0.0.1:99999")
			Expect(err).Should(HaveOccurred())
		})
		It("should fail for bad TCP port", func() {
			_, err := NewServer("127.0.0.1:99999", "127.0.0.1:1234")
			Expect(err).Should(HaveOccurred())
		})
	})
	Context("Start", func() {
		It("should work", func() {
			addr, _ := net.ResolveTCPAddr("tcp", "127.0.0.1")
			s := &Server{
				tcpAddr: addr,
				udpConn: uConn,
			}
			err := s.Start()
			Expect(err).Should(BeNil())
			uConn.EXPECT().Close()
			s.Stop()
		})
		It("should fail for bad IP/port", func() {
			addr, _ := net.ResolveTCPAddr("tcp", "123.123.123.123:123")
			s := &Server{
				tcpAddr: addr,
			}
			err := s.Start()
			Expect(err).Should(HaveOccurred())
		})
	})
	// Context("runServer", func() {
	// 	FIt("should work", func() {
	// 		s := &Server{
	// 			ch:          make(chan []byte, 1),
	// 			tcpListener: listener,
	// 		}
	// 		addr, _ := net.ResolveTCPAddr("tcp", "127.0.0.1:1234")
	// 		listener.EXPECT().Accept().Return(tcpConn, nil)
	// 		tcpConn.EXPECT().Read(gomock.Any()).Return(0, errors.New(""))
	// 		tcpConn.EXPECT().RemoteAddr().Return(addr)
	// 		tcpConn.EXPECT().Close()
	// 		s.runServer()
	// 	})
	// })
	Context("handleAPRS", func() {
		It("should work", func() {
			s := &Server{
				ch:      make(chan []byte, 1),
				udpConn: uConn,
			}
			uConn.EXPECT().Write(gomock.AssignableToTypeOf(make([]byte, 1))).Return(1, nil)

			go s.handleAPRS()
			s.ch <- testPyld
		})
		It("should fail for bad packet", func() {
			s := &Server{
				ch:      make(chan []byte, 1),
				udpConn: uConn,
			}

			go s.handleAPRS()
			s.ch <- []byte("123")
			time.Sleep(time.Millisecond * 10)
		})
		It("should have no position", func() {
			s := &Server{
				ch:      make(chan []byte, 1),
				udpConn: uConn,
			}

			go s.handleAPRS()
			s.ch <- []byte("KI4ABC>AB1CDE,WIDE2*,qAO,KI1ABC:@123456#-hello0000000000000000")
			time.Sleep(time.Millisecond * 10)
		})
		It("should fail on write", func() {
			s := &Server{
				ch:      make(chan []byte, 1),
				udpConn: uConn,
			}
			uConn.EXPECT().Write(gomock.AssignableToTypeOf(make([]byte, 1))).Return(0, errors.New("error"))

			go s.handleAPRS()
			s.ch <- testPyld
		})
	})
})
