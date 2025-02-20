package email

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/simpleauthlink/authapi/internal"
)

const (
	testServerAddr = "127.0.0.1"
	testServerPort = 2525
	testSenderName = "Test Sender"
	testSender     = "sender@testmail.com"
	testReceiver   = "receiver@testmail.com"
	testSubject    = "Test email"
	testBody       = "This is a test email"
	testHTMLBody   = "<h1>This is a test email</h1>"
)

var inboxChan = make(chan string, 1)

func TestMain(m *testing.M) {
	defer close(inboxChan)
	// create context with cancel
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// start test SMTP server to receive the email
	testSrv := internal.NewFakeSMTPServer(testServerAddr, testServerPort, inboxChan)
	if err := testSrv.Start(ctx); err != nil {
		panic(err)
	}
	defer testSrv.Stop()
	m.Run()
}

func TestValidEmail(t *testing.T) {
	if !(&Email{
		To:        testReceiver,
		Subject:   testSubject,
		Body:      nil,
		PlainBody: []byte(testBody),
	}).Valid() {
		t.Error("expected email to be valid")
	}
	if !(&Email{
		To:        testReceiver,
		Subject:   testSubject,
		Body:      []byte(testBody),
		PlainBody: nil,
	}).Valid() {
		t.Error("expected email to be valid")
	}
	if (&Email{
		To:        testReceiver,
		Subject:   "",
		Body:      nil,
		PlainBody: []byte(testBody),
	}).Valid() {
		t.Error("expected email to be invalid")
	}
	if (&Email{
		To:        "",
		Subject:   testSubject,
		Body:      nil,
		PlainBody: []byte(testBody),
	}).Valid() {
		t.Error("expected email to be invalid")
	}
	if (&Email{
		To:        "invalidEmail",
		Subject:   testSubject,
		Body:      nil,
		PlainBody: []byte(testBody),
	}).Valid() {
		t.Error("expected email to be invalid")
	}
	if (&Email{}).Valid() {
		t.Error("expected email to be invalid")
	}
}

func TestValidConfig(t *testing.T) {
	if !(&EmailConfig{
		SMTPServer:  testServerAddr,
		SMTPPort:    testServerPort,
		FromName:    testSenderName,
		FromAddress: testSender,
	}).Valid() {
		t.Error("expected config to be valid")
	}
	if (&EmailConfig{
		SMTPServer:  "",
		SMTPPort:    testServerPort,
		FromName:    testSenderName,
		FromAddress: testSender,
	}).Valid() {
		t.Error("expected config to be invalid")
	}
	if (&EmailConfig{
		SMTPServer:  testServerAddr,
		SMTPPort:    0,
		FromName:    testSenderName,
		FromAddress: testSender,
	}).Valid() {
		t.Error("expected config to be invalid")
	}
	if (&EmailConfig{
		SMTPServer:  testServerAddr,
		SMTPPort:    testServerPort,
		FromName:    "",
		FromAddress: testSender,
	}).Valid() {
		t.Error("expected config to be invalid")
	}
	if (&EmailConfig{
		SMTPServer:  testServerAddr,
		SMTPPort:    testServerPort,
		FromName:    testSenderName,
		FromAddress: "",
	}).Valid() {
		t.Error("expected config to be invalid")
	}
}

func TestNewEmailQueue(t *testing.T) {
	// create email queue with valid config
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	eq, err := NewEmailQueue(ctx, &EmailConfig{
		SMTPServer:  testServerAddr,
		SMTPPort:    testServerPort,
		FromName:    testSenderName,
		FromAddress: testSender,
	})
	if err != nil {
		t.Fatal(err)
	}
	if eq == nil {
		t.Error("expected email queue to be created")
	}
	// create email queue with auth
	eq, err = NewEmailQueue(ctx, &EmailConfig{
		SMTPServer:   testServerAddr,
		SMTPPort:     testServerPort,
		FromName:     testSenderName,
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
	eq, err = NewEmailQueue(ctx, &EmailConfig{
		SMTPServer:  "",
		SMTPPort:    testServerPort,
		FromName:    testSenderName,
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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	eq, err := NewEmailQueue(ctx, &EmailConfig{
		SMTPServer:  testServerAddr,
		SMTPPort:    testServerPort,
		FromName:    testSenderName,
		FromAddress: testSender,
	})
	if err != nil {
		t.Fatal(err)
	}
	// send email
	if err := eq.Send(Email{
		To:        testReceiver,
		Subject:   testSubject,
		Body:      []byte(testHTMLBody),
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
		if !strings.Contains(receivedMsg, testHTMLBody) {
			t.Errorf("expected email content to contain %q, got %q", testHTMLBody, receivedMsg)
		}
	case <-time.After(2 * time.Second):
		t.Error("timed out waiting for the email to be received")
	}
	// try to send invalid email
	if err := eq.Send(Email{}); err == nil {
		t.Error("expected error sending invalid email")
	}
	// try to compose a invalid email
	if body, err := eq.composeBody(Email{}); err == nil {
		t.Error("expected error composing invalid email")
	} else if body != nil {
		t.Error("expected body to be nil")
	}
	// try to send email to an invalid SMTP server
	badEq, err := NewEmailQueue(ctx, &EmailConfig{
		SMTPServer:  testServerAddr,
		SMTPPort:    8080,
		FromName:    testSenderName,
		FromAddress: testSender,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := badEq.Send(Email{
		To:        testReceiver,
		Subject:   testSubject,
		Body:      nil,
		PlainBody: []byte(testBody),
	}); err == nil {
		t.Error("expected error sending email")
	}
}

func TestPushSendEmail(t *testing.T) {
	// create email queue and start it
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	errCh := make(chan error, 1)
	eq, err := NewEmailQueue(ctx, &EmailConfig{
		SMTPServer:  testServerAddr,
		SMTPPort:    testServerPort,
		FromName:    testSenderName,
		FromAddress: testSender,
		ErrorCh:     errCh,
	})
	if err != nil {
		t.Fatal(err)
	}
	eq.Start()
	defer eq.Stop()
	// push email
	if err := eq.Push(Email{
		To:        testReceiver,
		Subject:   testSubject,
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
	// sleep to pop nil email
	time.Sleep(2 * time.Second)
	// push invalid email
	if err := eq.Push(Email{}); err == nil {
		t.Error("expected error pushing invalid email")
	}
}
