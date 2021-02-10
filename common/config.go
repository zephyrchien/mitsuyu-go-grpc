package common

import (
	"encoding/json"
	"io/ioutil"
)

func LoadServerConfig(file string, v *ServerConfig) error {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}
	return json.Unmarshal(content, v)
}

func LoadClientConfig(file string, v *ClientConfig) error {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}
	return json.Unmarshal(content, v)
}
