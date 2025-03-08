package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestResponseWith(t *testing.T) {
	type Data struct {
		Message string `json:"message"`
	}
	nilResp := ResponseWith[Data](nil)
	if !nilResp.empty {
		t.Errorf("expected response to be empty")
	}
	data := &Data{Message: "Hello, World!"}
	resp := ResponseWith(data)
	if resp.empty {
		t.Errorf("expected response to be non-empty")
	}
	if resp.Data.Message != data.Message {
		t.Errorf("expected %s, got %s", data.Message, resp.Data.Message)
	}
}

func TestOkResponse(t *testing.T) {
	resp := OkResponse()
	if !resp.empty {
		t.Errorf("expected response to be empty")
	}
}

func TestWriteJSON(t *testing.T) {
	type Data struct {
		Message string `json:"message"`
	}
	data := &Data{Message: "Hello, World!"}
	resp := ResponseWith(data)

	rr := httptest.NewRecorder()
	resp.WriteJSON(rr)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, status)
	}

	expected, _ := json.Marshal(data)
	if rr.Body.String() != string(expected)+"\n" {
		t.Errorf("expected body %s, got %s", string(expected), rr.Body.String())
	}
	// write json with nil data
	nilResp := ResponseWith[string](nil)
	rr = httptest.NewRecorder()
	nilResp.WriteJSON(rr)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, status)
	}

}

func TestWriteJSON_Empty(t *testing.T) {
	resp := OkResponse()

	rr := httptest.NewRecorder()
	resp.WriteJSON(rr)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, status)
	}

	if rr.Body.String() != "OK" {
		t.Errorf("expected body OK, got %s", rr.Body.String())
	}
}

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
