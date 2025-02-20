package internal

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"strings"
)

// FakeSMTPServer represents a simple SMTP testing server.
type FakeSMTPServer struct {
	addr     string
	inbox    chan string
	listener net.Listener
}

// NewFakeSMTPServer creates a new FakeSMTPServer instance that listens on the
// given address and port and stores the received emails in the inbox channel
// provided.
func NewFakeSMTPServer(addr string, port int, inbox chan string) *FakeSMTPServer {
	return &FakeSMTPServer{addr: fmt.Sprintf("%s:%d", addr, port), inbox: inbox}
}

// Start method launches the test SMTP server.
func (s *FakeSMTPServer) Start(ctx context.Context) error {
	var err error
	s.listener, err = net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	go func() {
		for {
			select {
			case <-ctx.Done():
				s.listener.Close()
			default:
				conn, err := s.listener.Accept()
				if err != nil {
					return
				}
				go s.handleConn(conn)
			}
		}
	}()
	return nil
}

// Stop method shuts down the test SMTP server.
func (s *FakeSMTPServer) Stop() {
	s.listener.Close()
}

func (s *FakeSMTPServer) handleConn(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	// send greeting
	fmt.Fprintf(conn, "220 Fake SMTP Service Ready\r\n")
	var dataBuilder strings.Builder
	inData := false
	// read incoming data
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return
		}
		line = strings.TrimRight(line, "\r\n")
		// check if we are in the data section
		if inData {
			if line == "." {
				inData = false
				// send back a confirmation and store the data
				fmt.Fprintf(conn, "250 OK\r\n")
				s.inbox <- dataBuilder.String()
				dataBuilder.Reset()
				continue
			}
			dataBuilder.WriteString(line + "\n")
			continue
		}
		// simple command handling
		switch {
		case strings.HasPrefix(line, "HELO"), strings.HasPrefix(line, "EHLO"):
			fmt.Fprintf(conn, "250 Hello\r\n")
		case strings.HasPrefix(line, "MAIL FROM:"):
			fmt.Fprintf(conn, "250 OK\r\n")
		case strings.HasPrefix(line, "RCPT TO:"):
			fmt.Fprintf(conn, "250 OK\r\n")
		case strings.HasPrefix(line, "DATA"):
			// prepare to receive data
			fmt.Fprintf(conn, "354 End data with <CR><LF>.<CR><LF>\r\n")
			inData = true
		case strings.HasPrefix(line, "QUIT"):
			// close the connection
			fmt.Fprintf(conn, "221 Bye\r\n")
			return
		default:
			fmt.Fprintf(conn, "250 OK\r\n")
		}
	}
}
