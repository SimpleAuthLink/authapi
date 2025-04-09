package osflag

import (
	"flag"
	"os"
	"testing"
	"time"
)

func resetCommandLine() {
	CommandLine = new(OsFlagSet)
	CommandLine.FlagSet = flag.NewFlagSet("", flag.ExitOnError)
	CommandLine.required = make(map[string]bool)

	// Filter out test framework flags
	os.Args = os.Args[:1]
}

func TestBoolVar(t *testing.T) {
	resetCommandLine()
	var flagValue bool
	os.Setenv("TEST_BOOL", "true")
	defer os.Unsetenv("TEST_BOOL")

	CommandLine.BoolVar(&flagValue, "TEST_BOOL", "boolFlag", false, "A boolean flag", false)
	if err := CommandLine.Parse(); err != nil {
		t.Fatalf("Failed to parse command line: %v", err)
	}

	if !flagValue {
		t.Errorf("Expected true, got %v", flagValue)
	}
}

func TestDurationVar(t *testing.T) {
	resetCommandLine()
	var flagValue time.Duration
	os.Setenv("TEST_DURATION", "5s")
	defer os.Unsetenv("TEST_DURATION")

	CommandLine.DurationVar(&flagValue, "TEST_DURATION", "durationFlag", 0, "A duration flag", false)
	if err := CommandLine.Parse(); err != nil {
		t.Fatalf("Failed to parse command line: %v", err)
	}

	if flagValue != 5*time.Second {
		t.Errorf("Expected 5s, got %v", flagValue)
	}
}

func TestFloat64Var(t *testing.T) {
	resetCommandLine()
	var flagValue float64
	os.Setenv("TEST_FLOAT", "3.14")
	defer os.Unsetenv("TEST_FLOAT")

	CommandLine.Float64Var(&flagValue, "TEST_FLOAT", "floatFlag", 0.0, "A float flag", false)
	if err := CommandLine.Parse(); err != nil {
		t.Fatalf("Failed to parse command line: %v", err)
	}

	if flagValue != 3.14 {
		t.Errorf("Expected 3.14, got %v", flagValue)
	}
}

func TestIntVar(t *testing.T) {
	resetCommandLine()
	var flagValue int
	os.Setenv("TEST_INT", "42")
	defer os.Unsetenv("TEST_INT")

	CommandLine.IntVar(&flagValue, "TEST_INT", "intFlag", 0, "An int flag", false)
	if err := CommandLine.Parse(); err != nil {
		t.Fatalf("Failed to parse command line: %v", err)
	}

	if flagValue != 42 {
		t.Errorf("Expected 42, got %v", flagValue)
	}
}

func TestStringVar(t *testing.T) {
	resetCommandLine()
	var flagValue string
	os.Setenv("TEST_STRING", "hello")
	defer os.Unsetenv("TEST_STRING")

	CommandLine.StringVar(&flagValue, "TEST_STRING", "stringFlag", "default", "A string flag", false)
	if err := CommandLine.Parse(); err != nil {
		t.Fatalf("Failed to parse command line: %v", err)
	}

	if flagValue != "hello" {
		t.Errorf("Expected 'hello', got %v", flagValue)
	}
}

func TestUintVar(t *testing.T) {
	resetCommandLine()
	var flagValue uint
	os.Setenv("TEST_UINT", "100")
	defer os.Unsetenv("TEST_UINT")

	CommandLine.UintVar(&flagValue, "TEST_UINT", "uintFlag", 0, "A uint flag", false)
	if err := CommandLine.Parse(); err != nil {
		t.Fatalf("Failed to parse command line: %v", err)
	}

	if flagValue != 100 {
		t.Errorf("Expected 100, got %v", flagValue)
	}
}

func TestRequiredFlag(t *testing.T) {
	resetCommandLine()
	var flagValue string
	CommandLine.StringVar(&flagValue, "", "requiredFlag", "", "A required flag", true)

	if err := CommandLine.Parse(); err == nil {
		t.Errorf("Expected error for missing required flag, got nil")
	}
}

func TestDefaultValues(t *testing.T) {
	resetCommandLine()
	var flagValue string
	CommandLine.StringVar(&flagValue, "", "defaultFlag", "defaultValue", "A flag with a default value", false)
	if err := CommandLine.Parse(); err != nil {
		t.Fatalf("Failed to parse command line: %v", err)
	}

	if flagValue != "defaultValue" {
		t.Errorf("Expected 'defaultValue', got %v", flagValue)
	}
}
