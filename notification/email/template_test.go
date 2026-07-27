package email

import (
	"testing"
)

var testTemplate = &EmailTemplate{
	HTML:  "<html><body><h1>{{.Title}}</h1><p>{{.Content}}</p></body></html>",
	Plain: "Title: {{.Title}}\nContent: {{.Content}}",
}

type testData struct {
	Title   string
	Content string
}

func TestCompose(t *testing.T) {
	// valid data
	data := testData{
		Title:   "Test Title",
		Content: "Test Content",
	}
	email, err := testTemplate.Compose(EmailParams{
		To:      testReceiver,
		Subject: testSubject,
	}, data)
	if err != nil {
		t.Fatalf("expected nil, got error: %v", err)
	}
	if email.Params.To != testReceiver {
		t.Fatalf("got %v, want %v", email.Params.To, testReceiver)
	}
	if email.Params.Subject != testSubject {
		t.Fatalf("got %v, want %v", email.Params.Subject, testSubject)
	}
	expectedBody := "<html><body><h1>Test Title</h1><p>Test Content</p></body></html>"
	expectedPlain := "Title: Test Title\nContent: Test Content"
	if string(email.Body) != expectedBody {
		t.Fatalf("got %v, want %v", string(email.Body), expectedBody)
	}
	if string(email.PlainBody) != expectedPlain {
		t.Fatalf("got %v, want %v", string(email.PlainBody), expectedPlain)
	}
	// no subject
	if _, err := testTemplate.Compose(EmailParams{
		To:      testReceiver,
		Subject: "",
	}, data); err == nil {
		t.Fatalf("expected error, got nil")
	}
	// no to address
	if _, err := testTemplate.Compose(EmailParams{
		To:      "",
		Subject: testSubject,
	}, data); err == nil {
		t.Fatalf("expected error, got nil")
	}
	// bad to address
	if _, err := testTemplate.Compose(EmailParams{
		To:      "bad email",
		Subject: testSubject,
	}, data); err == nil {
		t.Fatalf("expected error, got nil")
	}
	// invalid template
	emptyTemplate := &EmailTemplate{}
	validParams := EmailParams{
		To:      testReceiver,
		Subject: testSubject,
	}
	if _, err := emptyTemplate.Compose(validParams, data); err == nil {
		t.Fatalf("expected error, got nil")
	}
	// no html template
	noHTMLTemplate := &EmailTemplate{Plain: testTemplate.Plain}
	onlyPlainEmail, err := noHTMLTemplate.Compose(validParams, data)
	if err != nil {
		t.Fatalf("expected nil, got error: %v", err)
	}
	if onlyPlainEmail.Body != nil {
		t.Fatalf("expected nil, got %v", string(onlyPlainEmail.Body))
	}
	if string(onlyPlainEmail.PlainBody) != expectedPlain {
		t.Fatalf("got %v, want %v", string(onlyPlainEmail.PlainBody), expectedPlain)
	}
	// no plain template
	noPlainTemplate := &EmailTemplate{HTML: testTemplate.HTML}
	onlyHTMLEmail, err := noPlainTemplate.Compose(validParams, data)
	if err != nil {
		t.Fatalf("expected nil, got error: %v", err)
	}
	if string(onlyHTMLEmail.Body) != expectedBody {
		t.Fatalf("got %v, want %v", string(onlyHTMLEmail.Body), expectedBody)
	}
	if onlyHTMLEmail.PlainBody != nil {
		t.Fatalf("expected nil, got %v", string(onlyHTMLEmail.PlainBody))
	}
}

func Test_composePlain(t *testing.T) {
	// valid data and template
	data := testData{
		Title:   "Test Title",
		Content: "Test Content",
	}
	body, err := testTemplate.composePlain(data)
	if err != nil {
		t.Fatalf("expected nil, got error: %v", err)
	}
	expected := "Title: Test Title\nContent: Test Content"
	if string(body) != expected {
		t.Fatalf("got %v, want %v", string(body), expected)
	}
	// no plain template
	wrongPlainTemplate := *testTemplate
	wrongPlainTemplate.Plain = ""
	body, err = wrongPlainTemplate.composePlain(data)
	if err != nil {
		t.Fatalf("expected nil, got error: %v", err)
	}
	if body != nil {
		t.Fatalf("expected nil, got %v", string(body))
	}
}

func Test_composeHTML(t *testing.T) {
	// valid data and template
	data := testData{
		Title:   "Test Title",
		Content: "Test Content",
	}
	body, err := testTemplate.composeHTML(data)
	if err != nil {
		t.Fatalf("expected nil, got error: %v", err)
	}
	expected := "<html><body><h1>Test Title</h1><p>Test Content</p></body></html>"
	if string(body) != expected {
		t.Fatalf("got %v, want %v", string(body), expected)
	}
	// no html template
	wrongHTMLTemplate := *testTemplate
	wrongHTMLTemplate.HTML = ""
	body, err = wrongHTMLTemplate.composeHTML(data)
	if err != nil {
		t.Fatalf("expected nil, got error: %v", err)
	}
	if body != nil {
		t.Fatalf("expected nil, got %v", string(body))
	}
}

func TestComposeWithBrokenTemplate(t *testing.T) {
	validParams := EmailParams{
		To:      testReceiver,
		Subject: testSubject,
	}
	data := testData{Title: "T", Content: "C"}

	// Broken HTML template: Parse error
	t.Run("broken HTML template", func(t *testing.T) {
		bad := &EmailTemplate{
			HTML:  "{{.Title}", // missing closing brace
			Plain: "OK {{.Title}}",
		}
		_, err := bad.Compose(validParams, data)
		if err == nil {
			t.Error("expected parse error for broken HTML template")
		}
	})

	// Broken Plain template: Parse error
	t.Run("broken Plain template", func(t *testing.T) {
		bad := &EmailTemplate{
			HTML:  "OK {{.Title}}",
			Plain: "{{.Title}", // missing closing brace
		}
		_, err := bad.Compose(validParams, data)
		if err == nil {
			t.Error("expected parse error for broken Plain template")
		}
	})
}
