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
	CommandLine.flags = make(map[string]osflag)

	// Filter out test framework flags
	os.Args = os.Args[:1]
}

func TestBoolVar(t *testing.T) {
	resetCommandLine()
	var flagValue bool
	if err := os.Setenv("TEST_BOOL", "true"); err != nil {
		t.Fatalf("Failed to set env variable: %v", err)
	}
	defer func() {
		_ = os.Unsetenv("TEST_BOOL")
	}()

	CommandLine.BoolVar(&flagValue, "TEST_BOOL", "boolFlag", false, "A boolean flag", false)
	if err := CommandLine.Parse(nil); err != nil {
		t.Fatalf("Failed to parse command line: %v", err)
	}

	if !flagValue {
		t.Errorf("Expected true, got %v", flagValue)
	}
}

func TestDurationVar(t *testing.T) {
	resetCommandLine()
	var flagValue time.Duration
	if err := os.Setenv("TEST_DURATION", "5s"); err != nil {
		t.Fatalf("Failed to set env variable: %v", err)
	}
	defer func() {
		_ = os.Unsetenv("TEST_DURATION")
	}()

	CommandLine.DurationVar(&flagValue, "TEST_DURATION", "durationFlag", 0, "A duration flag", false)
	if err := CommandLine.Parse(nil); err != nil {
		t.Fatalf("Failed to parse command line: %v", err)
	}

	if flagValue != 5*time.Second {
		t.Errorf("Expected 5s, got %v", flagValue)
	}
}

func TestFloat64Var(t *testing.T) {
	resetCommandLine()
	var flagValue float64
	if err := os.Setenv("TEST_FLOAT", "3.14"); err != nil {
		t.Fatalf("Failed to set env variable: %v", err)
	}
	defer func() {
		_ = os.Unsetenv("TEST_FLOAT")
	}()

	CommandLine.Float64Var(&flagValue, "TEST_FLOAT", "floatFlag", 0.0, "A float flag", false)
	if err := CommandLine.Parse(nil); err != nil {
		t.Fatalf("Failed to parse command line: %v", err)
	}

	if flagValue != 3.14 {
		t.Errorf("Expected 3.14, got %v", flagValue)
	}
}

func TestIntVar(t *testing.T) {
	resetCommandLine()
	var flagValue int
	if err := os.Setenv("TEST_INT", "42"); err != nil {
		t.Fatalf("Failed to set env variable: %v", err)
	}
	defer func() {
		_ = os.Unsetenv("TEST_INT")
	}()

	CommandLine.IntVar(&flagValue, "TEST_INT", "intFlag", 0, "An int flag", false)
	if err := CommandLine.Parse(nil); err != nil {
		t.Fatalf("Failed to parse command line: %v", err)
	}

	if flagValue != 42 {
		t.Errorf("Expected 42, got %v", flagValue)
	}
}

func TestStringVar(t *testing.T) {
	resetCommandLine()
	var flagValue string
	if err := os.Setenv("TEST_STRING", "hello"); err != nil {
		t.Fatalf("Failed to set env variable: %v", err)
	}
	defer func() {
		_ = os.Unsetenv("TEST_STRING")
	}()

	CommandLine.StringVar(&flagValue, "TEST_STRING", "stringFlag", "default", "A string flag", false)
	if err := CommandLine.Parse(nil); err != nil {
		t.Fatalf("Failed to parse command line: %v", err)
	}

	if flagValue != "hello" {
		t.Errorf("Expected 'hello', got %v", flagValue)
	}
}

func TestUintVar(t *testing.T) {
	resetCommandLine()
	var flagValue uint
	if err := os.Setenv("TEST_UINT", "100"); err != nil {
		t.Fatalf("Failed to set env variable: %v", err)
	}
	defer func() {
		_ = os.Unsetenv("TEST_UINT")
	}()

	CommandLine.UintVar(&flagValue, "TEST_UINT", "uintFlag", 0, "A uint flag", false)
	if err := CommandLine.Parse(nil); err != nil {
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

	if err := CommandLine.Parse(nil); err == nil {
		t.Errorf("Expected error for missing required flag, got nil")
	}
}

func TestDefaultValues(t *testing.T) {
	resetCommandLine()
	var flagValue string
	CommandLine.StringVar(&flagValue, "", "defaultFlag", "defaultValue", "A flag with a default value", false)
	if err := CommandLine.Parse(nil); err != nil {
		t.Fatalf("Failed to parse command line: %v", err)
	}

	if flagValue != "defaultValue" {
		t.Errorf("Expected 'defaultValue', got %v", flagValue)
	}
}

func TestLoadEnv(t *testing.T) {
	resetCommandLine()
	// try to load a non-existing env file (should not error)
	if err := loadEnv("non_existing.env"); err != nil {
		t.Fatalf("Expected no error for non-existing env file, got: %v", err)
	}
	// create .env file
	envFileContent := []byte("TEST_ENV=envValue")
	envFilePath := ".env"
	if err := os.WriteFile(envFilePath, envFileContent, 0o644); err != nil {
		t.Fatalf("Failed to create env file: %v", err)
	}
	defer func() {
		_ = os.Remove(envFilePath)
	}()
	// parse flags and check the value
	var flagValue string
	CommandLine.StringVar(&flagValue, "TEST_ENV", "envFlag", "defaultValue", "A flag with an env variable", false)
	if err := CommandLine.Parse(nil); err != nil {
		t.Fatalf("Failed to parse command line: %v", err)
	}
	if flagValue != "envValue" {
		t.Errorf("Expected 'envValue', got %v", flagValue)
	}
}
