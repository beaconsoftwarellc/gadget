package environment

import (
	"os"
	"strconv"
)

// Get returns the value of an environment variable if the variable is set, else the default value
func Get(key string, defaultVal string) string {
	val, ok := os.LookupEnv(key)
	if !ok {
		val = defaultVal
	}
	return val
}

// RedisAddress returns the combined host and port from the environment
func RedisAddress() string {
	return RedisHost() + ":" + strconv.Itoa(RedisPort())
}

// RedisHost returns the host value from the environment
func RedisHost() string {
	return Get("PERSIST_HOST", "127.0.0.1")
}

// RedisPort returns the port value from the environment
func RedisPort() int {
	port, _ := strconv.Atoi(Get("PERSIST_PORT", "6379"))
	return port
}
