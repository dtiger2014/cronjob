package server

import (
	"encoding/json"
	"io/ioutil"
)

// Config : program config
type Config struct {
	Port    int    `json:"port"`
	WebRoot string `json:"webRoot"`
}

var (
	// GConfig : global config
	GConfig *Config
)

// InitConfig : init config file
func InitConfig(filename string) error {
	var (
		content []byte
		conf    Config
		err     error
	)

	// read file dir path
	if content, err = ioutil.ReadFile(filename); err != nil {
		return err
	}

	// json unmarshal
	if err = json.Unmarshal(content, &conf); err != nil {
		return err
	}

	GConfig = &conf

	return nil
}
