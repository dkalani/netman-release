package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type ServerConfig struct {
	ListenAddress string `json:"listen_address"`
}

func Unmarshal(input io.Reader) (*ServerConfig, error) {
	decoder := json.NewDecoder(input)

	c := &ServerConfig{}
	err := decoder.Decode(&c)
	if err != nil {
		return nil, fmt.Errorf("json decode: %s", err)
	}

	return c, nil
}

func (c *ServerConfig) Marshal(output io.Writer) error {
	encoder := json.NewEncoder(output)

	err := encoder.Encode(&c)
	if err != nil {
		return fmt.Errorf("json encode: %s", err) // not tested
	}

	return nil
}

func ParseConfigFile(configFilePath string) (*ServerConfig, error) {
	if configFilePath == "" {
		return nil, fmt.Errorf("missing config file path")
	}

	configFile, err := os.Open(configFilePath)
	if err != nil {
		return nil, err
	}
	defer configFile.Close()

	serverConfig, err := Unmarshal(configFile)
	if err != nil {
		return nil, fmt.Errorf("parsing config: %s", err)
	}

	return serverConfig, nil
}
