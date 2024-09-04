package config

import (
	"github.com/gookit/config/v2"
	"golang.org/x/time/rate"
)

type Config struct {
	Server Server `mapstructure:"server"`
	Log    Log    `mapstructure:"log"`
	MySql  MySql  `mapstructure:"mysql"`
	Redis  Redis  `mapstructure:"redis"`
}

type Server struct {
	MaxThreads int `mapstructure:"max_threads"`
	MaxTimeout int `mapstructure:"max_timeout"`
	RateLimit  int `mapstructure:"rate_limit"`
	rate.Limit
	Mode string `mapstructure:"mode"`
}

type Log struct {
	Level string `mapstructure:"level"`
	File  string `mapstructure:"file"`
}

type MySql struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Database string `mapstructure:"database"`
}

type Redis struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	Db       int    `mapstructure:"db"`
}

// New creates a new instance of the specified type T by loading the configuration from the given file(s).
// It takes a driver and one or more file paths as input and returns an instance of type T.
func New[T any](driver config.Driver, file ...string) (T, error) {
	// Add the driver to the configuration
	config.AddDriver(driver)

	// Create a new instance of type T
	var t T

	// Load the configuration from the file(s)
	if err := config.LoadFiles(file...); err != nil {
		return t, err
	}

	// Decode the configuration into the instance of type T
	if err := config.Decode(&t); err != nil {
		return t, err
	}

	// Return the created instance
	return t, nil
}

func NewFormDir[T any](driver config.Driver, dir string) (T, error) {
	// Add the driver to the configuration
	config.AddDriver(driver)

	// Create a new instance of type T
	var t T

	config.Reset()

	// Load the configuration from the file(s)
	if err := config.LoadFromDir(dir, driver.Name()); err != nil {
		return t, err
	}

	// Decode the configuration into the instance of type T
	if err := config.Decode(&t); err != nil {
		return t, err
	}

	// Return the created instance
	return t, nil
}
