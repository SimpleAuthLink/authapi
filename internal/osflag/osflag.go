package osflag

// osflag package provides a way to manage command line flags and environment
// variables in Go applications. It allows for the creation of command line
// flags that can also be overwritten by environment variables. By default it
// loads a `.env` file, but this can be overridden by passing an `WithEnvFile`
// option to the `Parse` method. It also checks for required flags and ensures
// that they are set before parsing the command line arguments.

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

// Options is a struct that holds options for the Parse method.
type Options struct {
	envFile string
}

// WithEnvFile function creates an Options instance with the specified
// envFile path. If the path is empty, it returns nil. This function is used
// to specify a custom env file path when calling the Parse method.
func WithEnvFile(path string) *Options {
	if path == "" {
		return nil
	}
	return &Options{envFile: path}
}

// osflag is a struct that holds the name, environment variable, and
// required mark of a flag. It is used to manage command line flags
// and their corresponding environment variables.
type osflag struct {
	name     string
	env      string
	required bool
}

// OsFlagSet is a struct that embeds flag.FlagSet and adds support for env
// variables. It allows for the creation of command line flags that can also
// be overwritten by environment variables. By default it loads `.env` file,
// but this can be overridden by passing an WithEnvFile option to the Parse
// method. It also checks for required flags and ensures that they are set
// before parsing the command line arguments.
type OsFlagSet struct {
	*flag.FlagSet
	flags  map[string]osflag
	parsed bool
}

// CommandLine is the default OsFlagSet instance.
var CommandLine *OsFlagSet

// init initializes the CommandLine variable with a new OsFlagSet instance.
func init() {
	CommandLine = new(OsFlagSet)
	if len(os.Args) == 0 {
		CommandLine.FlagSet = flag.NewFlagSet("", flag.ExitOnError)
	} else {
		CommandLine.FlagSet = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	}
	CommandLine.flags = make(map[string]osflag)
}

// BoolVar method registers a boolean flag with the given name, env variable,
// default value, usage string, and required mark.
func (of *OsFlagSet) BoolVar(p *bool, env, name string, value bool, usage string, required bool) {
	of.flags[name] = osflag{name, env, required}
	of.FlagSet.BoolVar(p, name, value, usage)
}

// DurationVar method registers a duration flag with the given name, env
// variable, default value, usage string, and required mark.
func (of *OsFlagSet) DurationVar(p *time.Duration, env, name string, value time.Duration, usage string, required bool) {
	of.flags[name] = osflag{name, env, required}
	of.FlagSet.DurationVar(p, name, value, usage)
}

// Float64Var method registers a float64 flag with the given name, env variable,
// default value, usage string, and required mark.
func (of *OsFlagSet) Float64Var(p *float64, env, name string, value float64, usage string, required bool) {
	of.flags[name] = osflag{name, env, required}
	of.FlagSet.Float64Var(p, name, value, usage)
}

// IntVar method registers an int flag with the given name, env variable,
// default value, usage string, and required mark.
func (of *OsFlagSet) IntVar(p *int, env, name string, value int, usage string, required bool) {
	of.flags[name] = osflag{name, env, required}
	of.FlagSet.IntVar(p, name, value, usage)
}

// StringVar method registers a string flag with the given name, env variable,
// default value, usage string, and required mark.
func (of *OsFlagSet) StringVar(p *string, env, name string, value string, usage string, required bool) {
	of.flags[name] = osflag{name, env, required}
	of.FlagSet.StringVar(p, name, value, usage)
}

// UintVar method registers a uint flag with the given name, env variable,
// default value, usage string, and required mark.
func (of *OsFlagSet) UintVar(p *uint, env, name string, value uint, usage string, required bool) {
	of.flags[name] = osflag{name, env, required}
	of.FlagSet.UintVar(p, name, value, usage)
}

// Parse method parses the command line arguments and loads the environment
// variables from the specified env file. It checks if all required flags are
// set and it overwrites the command line flags with the values from the env
// variables if they are set. It returns an error if any required flags are
// not set or if there is an error loading the env file.
func (of *OsFlagSet) Parse(opts *Options) error {
	if err := of.FlagSet.Parse(os.Args[1:]); err != nil {
		return err
	}
	// load the env file
	envFile := ".env"
	if opts != nil && opts.envFile != "" {
		envFile = opts.envFile
	}
	if err := loadEnv(envFile); err != nil {
		return fmt.Errorf("failed to load env file: %w", err)
	}
	// check if all required flags are set
	for name, osf := range of.flags {
		if envValue := os.Getenv(osf.env); envValue != "" {
			if err := of.FlagSet.Set(name, envValue); err != nil {
				return fmt.Errorf("failed to set flag %s from env: %w", name, err)
			}
		}
		// check if the flag is required and not set
		if osf.required {
			f := of.FlagSet.Lookup(name)
			if f == nil || f.Value.String() == "" {
				return fmt.Errorf("required flag %s is not set", name)
			}
		}
	}
	of.parsed = of.FlagSet.Parsed()
	return nil
}

// Parsed method returns true if the command line arguments have been parsed.
func (of *OsFlagSet) Parsed() bool {
	return of.parsed
}

// PrintDefaults method prints the default values of all flags.
func (of *OsFlagSet) PrintDefaults() {
	of.FlagSet.PrintDefaults()
}

// BoolVar method registers a boolean flag with the given name, env variable,
// default value, usage string, and required mark.
func BoolVar(p *bool, env, name string, value bool, usage string, required bool) {
	CommandLine.BoolVar(p, env, name, value, usage, required)
}

// DurationVar method registers a duration flag with the given name, env
// variable, default value, usage string, and required mark.
func DurationVar(p *time.Duration, env, name string, value time.Duration, usage string, required bool) {
	CommandLine.DurationVar(p, env, name, value, usage, required)
}

// Float64Var method registers a float64 flag with the given name, env variable,
// default value, usage string, and required mark.
func Float64Var(p *float64, env, name string, value float64, usage string, required bool) {
	CommandLine.Float64Var(p, env, name, value, usage, required)
}

// IntVar method registers an int flag with the given name, env variable,
// default value, usage string, and required mark.
func IntVar(p *int, env, name string, value int, usage string, required bool) {
	CommandLine.IntVar(p, env, name, value, usage, required)
}

// StringVar method registers a string flag with the given name, env variable,
// default value, usage string, and required mark.
func StringVar(p *string, env, name string, value string, usage string, required bool) {
	CommandLine.StringVar(p, env, name, value, usage, required)
}

// UintVar method registers a uint flag with the given name, env variable,
// default value, usage string, and required mark.
func UintVar(p *uint, env, name string, value uint, usage string, required bool) {
	CommandLine.UintVar(p, env, name, value, usage, required)
}

// Parse method parses the command line arguments and loads the environment
// variables from the specified env file. It checks if all required flags are
// set and it overwrites the command line flags with the values from the env
// variables if they are set. It returns an error if any required flags are
// not set or if there is an error loading the env file.
func Parse(opts *Options) error {
	return CommandLine.Parse(opts)
}

// Parsed method returns true if the command line arguments have been parsed.
func Parsed() bool {
	return CommandLine.parsed
}

// PrintDefaults method prints the default values of all flags.
func PrintDefaults() {
	CommandLine.PrintDefaults()
}

// loadEnv function loads environment variables from a file. If the file does
// not exist, it returns nil and does not raise an error. It reads the file
// line by line, ignoring empty lines and comments. It sets the environment
// variables in the current process using os.Setenv. It removes any "export "
// prefix and surrounding quotes from the variable assignments. It returns
// an error if there is an issue opening or reading the file (different from
// the file not existing).
func loadEnv(path string) error {
	envFile, err := os.Open(path)
	if err != nil {
		// if the file does not exist, return nil
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to open env file: %w", err)
	}
	defer envFile.Close()
	// create a line scanner
	scanner := bufio.NewScanner(envFile)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue // skip empty lines and comments
		}
		// remove "export " prefix if present
		line = strings.TrimPrefix(line, "export ")
		// split on the first '=' character
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue // or return an error if preferred
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		value = strings.Trim(value, "\"'") // remove surrounding quotes
		// set var in the current env
		os.Setenv(key, value)
	}
	return nil
}
