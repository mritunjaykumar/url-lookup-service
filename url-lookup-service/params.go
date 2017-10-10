package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type params struct {
	URL     string `json:"url,omitempty"`
	Mongodb string `json:"mongodb,omitempty"`
}

func (p *params) getParams(conf string) error {
	config, err := ioutil.ReadFile("./params.json")

	if err != nil {
		fmt.Printf("config error: %v\n", err)
		os.Exit(1)
	}

	err = json.Unmarshal(config, p)
	if err != nil {
		fmt.Printf("config error: %v\n", err)
		os.Exit(1)
	}

	return nil
}
