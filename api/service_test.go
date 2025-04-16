package api

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/simpleauthlink/authapi/internal/fakesmtpserver"
	"github.com/simpleauthlink/authapi/notification/email"
)

const (
	testServerAddr         = "127.0.0.1"
	testServerSMTPPort     = 2526
	testServerAPIPort      = 5555
	testServerSecret       = "server-secret"
	testSenderName         = "TestAPI"
	testSender             = "api-test@testmail.com"
	testUserEmail          = "user@testmail.com"
	testAppName            = "TestApp"
	testAppRedirectURL     = "http://testapp.com"
	testAppSessionDuration = time.Second * 35
	testAppSecret          = "test-secret"
)

var (
	testServerApiURL = fmt.Sprintf("http://%s:%d", testServerAddr, testServerAPIPort)
	inboxChan        = make(chan string, 1)
)

func TestMain(m *testing.M) {
	defer close(inboxChan)
	// create context with cancel
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// start test SMTP server to receive the email
	testSrv := fakesmtpserver.NewServer(testServerAddr, testServerSMTPPort, inboxChan)
	if err := testSrv.Start(ctx); err != nil {
		panic(err)
	}
	defer testSrv.Stop()
	// create email queue with valid config
	eq, err := email.NewEmailQueue(ctx, &email.EmailConfig{
		SMTPServer:  testServerAddr,
		SMTPPort:    testServerSMTPPort,
		FromName:    testSenderName,
		FromAddress: testSender,
	})
	if err != nil {
		panic(err)
	}
	eq.Start()
	defer eq.Stop()
	// create the API service
	apiSrv, err := New(ctx, &Config{
		Server:     testServerAddr,
		ServerPort: testServerAPIPort,
		Secret:     testServerSecret,
	}, eq)
	if err != nil {
		panic(err)
	}
	go func() {
		if err := apiSrv.Start(); err != nil {
			panic(err)
		}
	}()
	defer apiSrv.Stop()
	// make ping to the server to check if it is running
	nRetries := 5
	for {
		if nRetries == 0 {
			panic("API server is not running")
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
