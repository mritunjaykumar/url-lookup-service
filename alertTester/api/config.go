package api

import (
	"fmt"
	"encoding/json"
	"io/ioutil"
	"os"
)

type ApiConf struct {
	Host         			string `json:"Host,omitempty"`
	Port         			int
	KapacitorCLI 			string `json:"KapacitorCLI,omitempty"`
	UDPListenerPortForMetrics	int
}

var apiConfig *ApiConf
var KapacitorCliPath = ""

func GetPortForApi() int {
	SetConfig()
	return apiConfig.Port
}

func SetConfig() {
	if apiConfig == nil {
		file, e := ioutil.ReadFile("config.json")
		if e != nil {
			fmt.Printf("File error: %v\n", e)
			os.Exit(1)
		}

		apiConfig = new(ApiConf)
		err := json.Unmarshal(file, &apiConfig)

		if err != nil {
			fmt.Println("error:", err)
		}

		if apiConfig.KapacitorCLI != "" {
			KapacitorCliPath = apiConfig.KapacitorCLI
		}
	}
}