package io

import (
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
	if rr.Body.String() != string(expected)+"\n" { // Ensure newline is accounted for
		t.Errorf("expected body %s, got %s", string(expected)+"\n", rr.Body.String())
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

func TestOkResponse_WithBody(t *testing.T) {
	body := []byte("Hello, World!")
	resp := OkResponse(body...)

	if resp.empty {
		t.Errorf("expected response to be non-empty")
	}

	if string(resp.Data.([]byte)) != string(body) {
		t.Errorf("expected body %s, got %s", string(body), string(resp.Data.([]byte)))
	}
}

func TestWrite(t *testing.T) {
	data := []byte("Hello, World!")
	resp := ResponseWith(&data)

	rr := httptest.NewRecorder()
	resp.Write(rr)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, status)
	}

	// Adjust to match the expected plain text output
	if rr.Body.String() != string(data) {
		t.Errorf("expected body %s, got %s", string(data), rr.Body.String())
	}
}

func TestWrite_Empty(t *testing.T) {
	resp := OkResponse()

	rr := httptest.NewRecorder()
	resp.Write(rr)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, status)
	}

	if rr.Body.String() != "OK" {
		t.Errorf("expected body OK, got %s", rr.Body.String())
	}
}

func TestBytes(t *testing.T) {
	body := []byte("Hello, World!")
	resp := OkResponse(body...)

	data, ok := resp.bytes()
	if !ok {
		t.Errorf("expected bytes to be valid")
	}

	if string(data) != string(body) {
		t.Errorf("expected body %s, got %s", string(body), string(data))
	}
}

func TestBytes_Empty(t *testing.T) {
	resp := OkResponse()

	data, ok := resp.bytes()
	if !ok {
		t.Errorf("expected bytes to be valid")
	}
	if data != nil {
		t.Errorf("expected nil data for empty response, got %v", data)
	}

	// custom text response
	bmsg := []byte("Hello, World!")
	bresp := ResponseWith(&bmsg)
	bdata, ok := bresp.bytes()
	if !ok {
		t.Errorf("expected bytes to be valid")
	}
	if string(bdata) != string(bmsg) {
		t.Errorf("expected body %s, got %s", string(bmsg), string(bdata))
	}

	// string type
	msg := "Hello, World!"
	sresp := ResponseWith(&msg)
	sdata, ok := sresp.bytes()
	if !ok {
		t.Errorf("expected bytes to be valid")
	}
	if string(sdata) != msg {
		t.Errorf("expected body %s, got %s", msg, string(sdata))
	}

	// invalid type
	imsg := 123
	iresp := ResponseWith(&imsg)
	idata, ok := iresp.bytes()
	if ok {
		t.Errorf("expected bytes to be invalid")
	}
	if idata != nil {
		t.Errorf("expected nil data for invalid response, got %v", idata)
	}
}
