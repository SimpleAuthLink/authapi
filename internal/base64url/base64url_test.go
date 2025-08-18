package base64url

import (
	"bytes"
	"testing"
)

func TestRawEncode(t *testing.T) {
	inputs := [][]byte{
		// normal
		[]byte("Hello, World!"),
		// some input that produce "+"
		{0xFF, 0xFF, 0xFF},
		// some input that produce "/"
		{0xFB, 0xEF, 0xBF},
	}
	expectedOutputs := [][]byte{
		[]byte("SGVsbG8sIFdvcmxkIQ"),
		[]byte("____"),
		[]byte("---_"),
	}
	for i, input := range inputs {
		encoded := RawEncode(input)
		if !bytes.Equal(encoded, expectedOutputs[i]) {
			t.Errorf("Unexpected output for input %q: got %q, want %q",
				input, encoded, expectedOutputs[i])
		}
	}
}

func TestRawDecode(t *testing.T) {
	t.Run("success", func(t *testing.T) {

		inputs := [][]byte{
			// normal
			[]byte("SGVsbG8sIFdvcmxkIQ"),
			// some input that produce "+"
			[]byte("____"),
			// some input that produce "/"
			[]byte("---_"),
		}
		expectedOutputs := [][]byte{
			[]byte("Hello, World!"),
			{0xFF, 0xFF, 0xFF},
			{0xFB, 0xEF, 0xBF},
		}
		for i, input := range inputs {
			decoded, err := RawDecode(input)
			if err != nil {
				t.Errorf("Unexpected error for input %q: %v",
					input, err)
				continue
			}
			if !bytes.Equal(decoded, expectedOutputs[i]) {
				t.Errorf("Unexpected output for input %q: got %q, want %q",
					input, decoded, expectedOutputs[i])
			}
		}
	})
	t.Run("decode error", func(t *testing.T) {
		invalidInputs := [][]byte{
			[]byte("SGVsbG8sIFdvcmxkIQ==="),
		}
		for _, input := range invalidInputs {
			decoded, err := RawDecode(input)
			if err == nil {
				t.Errorf("Expected error for input %q, got %q",
					input, decoded)
			}
		}
	})
}

func TestRawEncodeDecode(t *testing.T) {
	inputs := [][]byte{
		// normal
		[]byte("Hello, World!"),
		// some input that produce "+"
		{0xFF, 0xFF, 0xFF},
		// some input that produce "/"
		{0xFB, 0xEF, 0xBF},
	}
	for _, input := range inputs {
		encoded := RawEncode(input)
		decoded, err := RawDecode(encoded)
		if err != nil {
			t.Errorf("Unexpected error for input %q: %v",
				input, err)
			continue
		}
		if !bytes.Equal(decoded, input) {
			t.Errorf("Unexpected output for input %q: got %q, want %q",
				input, decoded, input)
		}
	}
}
