package env

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Utility methods for getting environment variables.
// The methods are prefixed with 'Must' because they will panic when smth goes wrong - typically
// used for halting startup of an application.
// For other scenarios one can create functions that will always return a value, either the found
// one or the default. These functions would not be prefixed with 'Must'.

func MustGetStringOrDefault(key, defaultVal string) string {
	val, present := os.LookupEnv(key)
	if !present {
		return defaultVal
	}
	val = strings.TrimSpace(val)
	if val == "" {
		return defaultVal
	}
	return val
}

func MustGetDurationOrDefault(key string, defaultVal time.Duration) time.Duration {
	val, present := os.LookupEnv(key)
	if !present {
		return defaultVal
	}
	val = strings.TrimSpace(val)
	if val == "" {
		panic(fmt.Sprintf("env var %s is provided but empty", key))
	}

	dur, err := time.ParseDuration(val)
	if err != nil {
		panic(fmt.Sprintf("env var %s is not a duration: %q", key, val))
	}

	return dur
}

func MustGetIntOrDefault(key string, defaultVal int64) int64 {
	val, present := os.LookupEnv(key)
	if !present {
		return defaultVal
	}
	val = strings.TrimSpace(val)
	if val == "" {
		panic(fmt.Sprintf("env var %s is provided but empty", key))
	}

	intVal, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		panic(fmt.Sprintf("env var %s is not an int: %q", key, val))
	}

	return intVal
}

func MustGetInt(key string) int64 {
	val, present := os.LookupEnv(key)
	if !present {
		panic(fmt.Sprintf("env var %s is missing", key))
	}
	val = strings.TrimSpace(val)
	if val == "" {
		panic(fmt.Sprintf("env var %s is provided but empty", key))
	}

	intVal, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		panic(fmt.Sprintf("env var %s is not an int: %q", key, val))
	}

	return intVal
}

func MustGetString(key string) string {
	val, present := os.LookupEnv(key)
	if !present {
		panic(fmt.Sprintf("env var %s is missing", key))
	}
	val = strings.TrimSpace(val)
	if val == "" {
		panic(fmt.Sprintf("env var %s is provided but empty", key))
	}
	return val
}
