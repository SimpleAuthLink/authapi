package authapi

import (
	"context"
	"testing"
	"time"

	"github.com/simpleauthlink/authapi/db"
	"github.com/simpleauthlink/authapi/email"
)

func TestNew(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	testDB := new(db.TempDriver)
	testDB.Init(nil)
	srv, err := New(ctx, testDB, &Config{
		Server:          "localhost",
		ServerPort:      8080,
		CleanerCooldown: 30 * time.Second,
		EmailConfig: email.EmailConfig{
			EmailHost: "smtp.gmail.com",
			EmailPort: 587,
			Address:   "",
			Password:  "",
		},
	})
	if err != nil {
		t.Errorf("expected nil, got %v", err)
		return
	}
	if srv == nil {
		t.Errorf("expected not nil, got nil")
		return
	}
}
