package conf

import (
	"encoding/json"
	"io"
	"io/ioutil"
)

func LoadJson(reader io.Reader, configData interface{}) error {
	buf, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(buf, configData); err != nil {
		return err
	}

	return nil
}

func LoadJsonFromFile(confFile string, configData interface{}) error {
	buf, err := ioutil.ReadFile(confFile)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(buf, configData); err != nil {
		return err
	}

	return nil
}
