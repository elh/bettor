// Package envflag provides a wrapper around the standard flag package that can fall back to environment variables
package envflag

import (
	"flag"
	"os"
	"strconv"
)

// Int defines an int flag with specified name, default value, and usage string.  If an environment variable with the
// same name as the flag is set, it will be used instead of the default value. The return value is the address of an int
// variable that stores the value of the flag.
func Int(name string, value int, usage string) *int {
	return flag.Int(name, envInt(name, value), usage)
}

// String defines a string flag with specified name, default value, and usage string. If an environment variable with
// the same name as the flag is set, it will be used instead of the default value. The return value is the address of a
// string variable that stores the value of the flag.
func String(name string, value string, usage string) *string {
	return flag.String(name, envString(name, value), usage)
}

// Parse parses the command-line flags from os.Args[1:]. Must be called after all flags are defined and before flags are
// accessed by the program.
func Parse() {
	flag.Parse()
}

func envInt(key string, defaultVal int) int {
	if val, ok := os.LookupEnv(key); ok {
		v, err := strconv.Atoi(val)
		if err == nil {
			return v
		}
	}
	return defaultVal
}

func envString(key string, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal
}
