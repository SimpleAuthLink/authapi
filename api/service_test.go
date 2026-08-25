package api

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/simpleauthlink/authapi/notification"
	"github.com/simpleauthlink/authapi/notification/email"
	xnet "go.k7z7z.cc/x/net"
	"go.k7z7z.cc/x/net/smtp/testsmtp"
)

var (
	testServerSMTPPort int
	testServerAPIPort  int
	testServerAPIURL   string
	inboxChan          = make(chan string, 1)
)

const (
	testServerAddr         = "127.0.0.1"
	testServerSecret       = "server-secret"
	testSender             = "api-test@testmail.com"
	testUserEmail          = "user@testmail.com"
	testAppName            = "TestApp"
	testAppRedirectURL     = "http://testapp.com"
	testAppSessionDuration = time.Second * 35
	testAppSecret          = "test-secret"
)

func TestMain(m *testing.M) {
	var err error
	testServerSMTPPort, err = xnet.SafeTestPort()
	if err != nil {
		log.Fatal(err)
	}
	testServerAPIPort, err = xnet.SafeTestPort()
	if err != nil {
		log.Fatal(err)
	}
	testServerAPIURL = fmt.Sprintf("http://%s:%d", testServerAddr, testServerAPIPort)
	defer close(inboxChan)
	// create context with cancel
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// start test SMTP server to receive the email
	testSrv := testsmtp.NewServer(testServerAddr, testServerSMTPPort, inboxChan)
	if err := testSrv.Start(ctx); err != nil {
		log.Fatal(err)
	}
	defer testSrv.Stop()
	// create email queue with valid config
	eq, err := notification.NewQueue(ctx, 10, &email.EmailConfig{
		SMTPServer:  testServerAddr,
		SMTPPort:    testServerSMTPPort,
		FromAddress: testSender,
	})
	if err != nil {
		log.Fatal(err)
	}
	eq.Start(1)
	defer eq.Stop()
	// create the API service
	apiSrv, err := New(ctx, &Config{
		Server:     testServerAddr,
		ServerPort: testServerAPIPort,
		Secret:     testServerSecret,
	}, eq)
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		if err := apiSrv.Start(); err != nil {
			log.Fatal(err)
		}
	}()
	defer func() {
		if err := apiSrv.Stop(); err != nil {
			log.Fatal(err)
		}
	}()
	// make ping to the server to check if it is running
	nRetries := 5
	for {
		if nRetries == 0 {
			log.Fatal("API server is not running")
		}
		if ok := apiSrv.Ping(); ok {
			break
		}
		nRetries--
		time.Sleep(time.Second)
	}
	// run the tests
	os.Exit(m.Run())
}

func TestStartError(t *testing.T) {
	port, err := xnet.SafeTestPort()
	if err != nil {
		t.Fatal(err)
	}
	// Block the port so ListenAndServe fails
	l, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = l.Close() }()

	eq, err := notification.NewQueue(t.Context(), 10)
	if err != nil {
		t.Fatal(err)
	}

	srv, err := New(t.Context(), &Config{
		Server:     testServerAddr,
		ServerPort: port,
		Secret:     testServerSecret,
	}, eq)
	if err != nil {
		t.Fatal(err)
	}

	err = srv.Start()
	if err == nil {
		t.Error("expected error when port is in use")
	}
}

func TestServiceStop(t *testing.T) {
	tempServerAPIPort, err := xnet.SafeTestPort()
	if err != nil {
		t.Fatal(err)
	}
	tempServerSMTPPort, err := xnet.SafeTestPort()
	if err != nil {
		t.Fatal(err)
	}
	// create email queue with valid config
	eq, err := notification.NewQueue(t.Context(), 10, &email.EmailConfig{
		SMTPServer:  testServerAddr,
		SMTPPort:    tempServerSMTPPort,
		FromAddress: testSender,
	})
	if err != nil {
		log.Fatal(err)
	}
	eq.Start(1)
	defer eq.Stop()
	// create the API service
	tempSrv, err := New(t.Context(), &Config{
		Server:     testServerAddr,
		ServerPort: tempServerAPIPort,
		Secret:     testServerSecret,
	}, eq)
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		if err := tempSrv.Start(); err != nil {
			log.Fatal(err)
		}
	}()
	// wait for the server to be up before pinging
	for range 10 {
		if tempSrv.Ping() {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	// verify server is up
	if !tempSrv.Ping() {
		t.Error("server should be reachable after start")
	}

	if err := tempSrv.Stop(); err != nil {
		t.Fatalf("stop returned error: %v", err)
	}
	// verify server is down
	if tempSrv.Ping() {
		t.Error("server should not be reachable after stop")
	}
}

func TestPingInvalidConfig(t *testing.T) {
	srv := &APIService{cfg: &Config{Server: "invalid host"}}
	if srv.Ping() {
		t.Error("expected Ping to return false for an invalid server address")
	}
}

type errorOnCloseListener struct {
	net.Listener
	started   chan struct{}
	startOnce sync.Once
}

func (l *errorOnCloseListener) Accept() (net.Conn, error) {
	l.startOnce.Do(func() { close(l.started) })
	return l.Listener.Accept()
}

func (l *errorOnCloseListener) Close() error {
	_ = l.Listener.Close()
	return fmt.Errorf("forced close error")
}

func TestStopCloseError(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	errL := &errorOnCloseListener{Listener: l, started: make(chan struct{})}
	ctx, cancel := context.WithCancel(context.Background())
	srv := &APIService{
		ctx:    ctx,
		cancel: cancel,
		cfg:    &Config{Server: testServerAddr},
		Server: &http.Server{},
	}
	go func() { _ = srv.Server.Serve(errL) }()
	<-errL.started
	if err := srv.Stop(); err == nil {
		t.Error("expected Stop to return the listener close error")
	}
}
