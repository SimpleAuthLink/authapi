package email

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/simpleauthlink/authapi/notification"
	xnet "go.k7z7z.cc/x/net"
	"go.k7z7z.cc/x/net/smtp/testsmtp"
)

const (
	testServerAddr = "127.0.0.1"
	testSender     = "sender@testmail.com"
	testReceiver   = "receiver@testmail.com"
	testSubject    = "Test email"
	testBody       = "This is a test email"
	testHTMLBody   = "<h1>This is a test email</h1>"
)

var (
	testServerPort int
	inboxChan      = make(chan string, 1)
)

func TestMain(m *testing.M) {
	var err error
	if testServerPort, err = xnet.SafeTestPort(); err != nil {
		panic(err)
	}
	defer close(inboxChan)
	// create context with cancel
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// start test SMTP server to receive the email
	testSrv := testsmtp.NewServer(testServerAddr, testServerPort, inboxChan)
	if err := testSrv.Start(ctx); err != nil {
		panic(err)
	}
	defer testSrv.Stop()
	os.Exit(m.Run())
}

func TestValidEmail(t *testing.T) {
	if !(&EmailNotification{
		Params: EmailParams{
			To:      testReceiver,
			Subject: testSubject,
		},
		Body:      nil,
		PlainBody: []byte(testBody),
	}).Valid() {
		t.Error("expected email to be valid")
	}
	if !(&EmailNotification{
		Params: EmailParams{
			To:      testReceiver,
			Subject: testSubject,
		},
		Body:      []byte(testBody),
		PlainBody: nil,
	}).Valid() {
		t.Error("expected email to be valid")
	}
	if (&EmailNotification{
		Params: EmailParams{
			To:      testReceiver,
			Subject: "",
		},
		Body:      nil,
		PlainBody: []byte(testBody),
	}).Valid() {
		t.Error("expected email to be invalid")
	}
	if (&EmailNotification{
		Params: EmailParams{
			To:      "",
			Subject: testSubject,
		},
		Body:      nil,
		PlainBody: []byte(testBody),
	}).Valid() {
		t.Error("expected email to be invalid")
	}
	if (&EmailNotification{
		Params: EmailParams{
			To:      "invalidEmail",
			Subject: testSubject,
		},
		Body:      nil,
		PlainBody: []byte(testBody),
	}).Valid() {
		t.Error("expected email to be invalid")
	}
	if (&EmailNotification{}).Valid() {
		t.Error("expected email to be invalid")
	}
}

func TestValidConfig(t *testing.T) {
	if !(&EmailConfig{
		SMTPServer:  testServerAddr,
		SMTPPort:    testServerPort,
		FromAddress: testSender,
	}).Valid() {
		t.Error("expected config to be valid")
	}
	if (&EmailConfig{
		SMTPServer:  "",
		SMTPPort:    testServerPort,
		FromAddress: testSender,
	}).Valid() {
		t.Error("expected config to be invalid")
	}
	if (&EmailConfig{
		SMTPServer:  testServerAddr,
		SMTPPort:    0,
		FromAddress: testSender,
	}).Valid() {
		t.Error("expected config to be invalid")
	}
	if (&EmailConfig{
		SMTPServer:  testServerAddr,
		SMTPPort:    testServerPort,
		FromAddress: "",
	}).Valid() {
		t.Error("expected config to be invalid")
	}
}

func TestNewEmailQueue(t *testing.T) {
	// create email queue with valid config
	eq, err := notification.NewQueue(t.Context(), 10, &EmailConfig{
		SMTPServer:   testServerAddr,
		SMTPPort:     testServerPort,
		FromAddress:  testSender,
		SMTPUsername: "username",
		SMTPPassword: "password",
	})
	if err != nil {
		t.Fatal(err)
	}
	if eq == nil {
		t.Error("expected email queue to be created")
	}
	// create email queue with invalid config
	eq, err = notification.NewQueue(t.Context(), 10, &EmailConfig{
		SMTPServer:  "",
		SMTPPort:    testServerPort,
		FromAddress: testSender,
	})
	if err == nil {
		t.Error("expected error creating email queue")
	}
	if eq != nil {
		t.Error("expected email queue to be nil")
	}
}

func TestSendEmail(t *testing.T) {
	// create email queue but don't start it
	config := &EmailConfig{
		SMTPServer:  testServerAddr,
		SMTPPort:    testServerPort,
		FromAddress: testSender,
	}
	eq, err := notification.NewQueue(t.Context(), 10, config)
	if err != nil {
		t.Fatal(err)
	}
	eq.Start(1)
	defer eq.Stop()
	// send email
	emailNotification := &EmailNotification{
		Params: EmailParams{
			To:      testReceiver,
			Subject: testSubject,
		},
		Body:      []byte(testHTMLBody),
		PlainBody: []byte(testBody),
	}
	if err := emailNotification.Send(config); err != nil {
		t.Fatal(err)
	}
	// check if the email was received
	select {
	case receivedMsg := <-inboxChan:
		if !strings.Contains(receivedMsg, testSubject) {
			t.Errorf("expected email content to contain %q, got %q", testSubject, receivedMsg)
		}
		if !strings.Contains(receivedMsg, testBody) {
			t.Errorf("expected email content to contain %q, got %q", testBody, receivedMsg)
		}
		if !strings.Contains(receivedMsg, testHTMLBody) {
			t.Errorf("expected email content to contain %q, got %q", testHTMLBody, receivedMsg)
		}
	case <-time.After(2 * time.Second):
		t.Error("timed out waiting for the email to be received")
	}
	invalidEmail := &EmailNotification{}
	// try to send invalid email
	if err := invalidEmail.Send(config); err == nil {
		t.Error("expected error sending invalid email")
	}
	// try to compose a invalid email
	if body, err := invalidEmail.composeBody(config); err == nil {
		t.Error("expected error composing invalid email")
	} else if body != nil {
		t.Error("expected body to be nil")
	}
	// try to send email to an invalid SMTP server
	badConf := &EmailConfig{
		SMTPServer:   testServerAddr,
		SMTPPort:     8080,
		FromAddress:  testSender,
		SMTPUsername: "user",
		SMTPPassword: "pass",
	}
	badEq, err := notification.NewQueue(t.Context(), 10, badConf)
	if err != nil {
		t.Fatal(err)
	}
	badEq.Start(1)
	defer badEq.Stop()
	otherNotification := &EmailNotification{
		Params: EmailParams{
			To:      testReceiver,
			Subject: testSubject,
		},
		Body:      nil,
		PlainBody: []byte(testBody),
	}
	if err := otherNotification.Send(badConf); err == nil {
		t.Error("expected error sending email")
	}
}

func TestPushSendEmail(t *testing.T) {
	// create email queue and start it
	eq, err := notification.NewQueue(t.Context(), 10, &EmailConfig{
		SMTPServer:  testServerAddr,
		SMTPPort:    testServerPort,
		FromAddress: testSender,
	})
	if err != nil {
		t.Fatal(err)
	}
	eq.Start(1)
	defer eq.Stop()
	// push email
	if err := eq.Push(&EmailNotification{
		Params: EmailParams{
			To:      testReceiver,
			Subject: testSubject,
		},
		Body:      nil,
		PlainBody: []byte(testBody),
	}); err != nil {
		t.Fatal(err)
	}
	// check if the email was received
	select {
	case receivedMsg := <-inboxChan:
		if !strings.Contains(receivedMsg, testSubject) {
			t.Errorf("expected email content to contain %q, got %q", testSubject, receivedMsg)
		}
		if !strings.Contains(receivedMsg, testBody) {
			t.Errorf("expected email content to contain %q, got %q", testBody, receivedMsg)
		}
	case <-time.After(2 * time.Second):
		t.Error("timed out waiting for the email to be received")
	}
	// push invalid email
	if err := eq.Push(&EmailNotification{}); err == nil {
		t.Error("expected error pushing invalid email")
	}
}

func TestSendWrongConfigType(t *testing.T) {
	n := &EmailNotification{
		Params:    EmailParams{To: testReceiver, Subject: testSubject},
		PlainBody: []byte(testBody),
	}
	// Pass a non-EmailConfig to trigger the type assertion failure
	if err := n.Send(nil); err != ErrInvalidConfig {
		t.Errorf("expected ErrInvalidConfig, got %v", err)
	}
}
