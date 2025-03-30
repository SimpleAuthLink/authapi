package osflag

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"
)

type OsFlagSet struct {
	*flag.FlagSet
	required map[string]bool
	parsed   bool
}

var CommandLine *OsFlagSet

func init() {
	CommandLine = new(OsFlagSet)
	if len(os.Args) == 0 {
		CommandLine.FlagSet = flag.NewFlagSet("", flag.ExitOnError)
	} else {
		CommandLine.FlagSet = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	}
	CommandLine.required = make(map[string]bool)
}

func (of *OsFlagSet) BoolVar(p *bool, env, name string, value bool, usage string, required bool) {
	var newDefault bool = value
	if rawBool := os.Getenv(env); rawBool != "" {
		if rawBool == "true" || rawBool == "True" || rawBool == "TRUE" || rawBool == "1" {
			newDefault = true
		}
	}
	of.required[name] = required
	of.FlagSet.BoolVar(p, name, newDefault, usage)
}

func (of *OsFlagSet) DurationVar(p *time.Duration, env, name string, value time.Duration, usage string, required bool) {
	var newDefault time.Duration = value
	if rawDuration := os.Getenv(env); rawDuration != "" {
		if dur, err := time.ParseDuration(rawDuration); err == nil {
			newDefault = dur
		}
	}
	of.required[name] = required
	of.FlagSet.DurationVar(p, name, newDefault, usage)
}

func (of *OsFlagSet) Float64Var(p *float64, env, name string, value float64, usage string, required bool) {
	var newDefault float64 = value
	if rawFloat := os.Getenv(env); rawFloat != "" {
		if f, err := strconv.ParseFloat(rawFloat, 64); err == nil {
			newDefault = f
		}
	}
	of.required[name] = required
	of.FlagSet.Float64Var(p, name, newDefault, usage)
}

func (of *OsFlagSet) IntVar(p *int, env, name string, value int, usage string, required bool) {
	var newDefault int = value
	if rawInt := os.Getenv(env); rawInt != "" {
		if integer, err := strconv.Atoi(rawInt); err == nil {
			newDefault = integer
		}
	}
	of.required[name] = required
	of.FlagSet.IntVar(p, name, newDefault, usage)
}

func (of *OsFlagSet) StringVar(p *string, env, name string, value string, usage string, required bool) {
	var newDefault string = value
	if rawString := os.Getenv(env); rawString != "" {
		newDefault = rawString
	}
	of.required[name] = required
	of.FlagSet.StringVar(p, name, newDefault, usage)
}

func (of *OsFlagSet) UintVar(p *uint, env, name string, value uint, usage string, required bool) {
	var newDefault uint
	if rawUint := os.Getenv(env); rawUint != "" {
		if ui, err := strconv.ParseUint(rawUint, 10, 64); err == nil {
			newDefault = uint(ui)
		}
	}
	of.required[name] = required
	of.FlagSet.UintVar(p, name, newDefault, usage)
}

func (of *OsFlagSet) Parse() error {
	if err := of.FlagSet.Parse(os.Args[1:]); err != nil {
		return err
	}
	// check if all required flags are set
	for name, required := range of.required {
		if required {
			f := of.FlagSet.Lookup(name)
			if f == nil || f.Value.String() == "" {
				return fmt.Errorf("required flag %s is not set", name)
			}
		}
	}
	of.parsed = of.FlagSet.Parsed()
	return nil
}

func (of *OsFlagSet) Parsed() bool {
	return of.parsed
}

func (of *OsFlagSet) PrintDefaults() {
	of.FlagSet.PrintDefaults()
}

func BoolVar(p *bool, env, name string, value bool, usage string, required bool) {
	CommandLine.BoolVar(p, env, name, value, usage, required)
}

func DurationVar(p *time.Duration, env, name string, value time.Duration, usage string, required bool) {
	CommandLine.DurationVar(p, env, name, value, usage, required)
}

func Float64Var(p *float64, env, name string, value float64, usage string, required bool) {
	CommandLine.Float64Var(p, env, name, value, usage, required)
}

func IntVar(p *int, env, name string, value int, usage string, required bool) {
	CommandLine.IntVar(p, env, name, value, usage, required)
}

func StringVar(p *string, env, name string, value string, usage string, required bool) {
	CommandLine.StringVar(p, env, name, value, usage, required)
}

func UintVar(p *uint, env, name string, value uint, usage string, required bool) {
	CommandLine.UintVar(p, env, name, value, usage, required)
}

func Parse() error {
	return CommandLine.Parse()
}

func Parsed() bool {
	return CommandLine.parsed
}

func PrintDefaults() {
	CommandLine.PrintDefaults()
}
