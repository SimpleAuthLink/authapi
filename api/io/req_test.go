package io

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
)

func TestRead(t *testing.T) {
	type Data struct {
		Message string `json:"message"`
	}
	data := &Data{Message: "Hello, World!"}
	body, _ := json.Marshal(data)
	req, err := http.NewRequest("POST", "/", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}

	var request Request[Data]
	if err := request.Read(req); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if request.Data.Message != data.Message {
		t.Errorf("expected %s, got %s", data.Message, request.Data.Message)
	}
}

func TestRead_EmptyBody(t *testing.T) {
	noBody, err := http.NewRequest("POST", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	if nilReq := new(Request[any]).Read(noBody); nilReq == nil {
		t.Errorf("expected error, got nil")
	}

	req, err := http.NewRequest("POST", "/", bytes.NewBuffer([]byte("")))
	if err != nil {
		t.Fatal(err)
	}

	var request *Request[any]
	err = request.Read(req)
	if err == nil {
		t.Errorf("expected error, got nil")
	}
	err = new(Request[any]).Read(req)
	if err == nil {
		t.Errorf("expected error, got nil")
	}
	if err.Error() != "empty request body" {
		t.Errorf("expected empty request body error, got %v", err)
	}
}
