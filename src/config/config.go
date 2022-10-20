package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

var Data = make(map[string]string)
var requiredConfigs = [10]string{"HOST", "PORT", "PERSISTENCE", "LOG_TARGET", "LOG_PATH", "LOG_LEVEL", "CORS_HEADER", "CORS_ORIGIN", "SSL_CERT_FILE", "SSL_KEY_FILE"}

func Init() {
	// first lets check if there is a parseable config file
	handleConfigFile()

	// now we try to get the params from env, priority env > config
	handleEnv()
}

func GetValue(key string) string {
	val, exist := Data[key]
	if !exist {
		fmt.Printf("> Missing config %s exiting server.", key)
		os.Exit(0)
	}
	return val
}

func handleEnv() {
	// check the env for the required configs and overwrite/write in our Data map
	for _, name := range requiredConfigs {
		value := os.Getenv(name)
		if value != "" {
			Data[name] = value
		}
	}
}

func handleConfigFile() {
	// first we read the json data
	data, err := ioutil.ReadFile("config.json")
	if nil != err {
		fmt.Print("> Config file could not be found or is not readable")
		os.Exit(0)
		return
	}
	// now we parse the config contents
	// lets see if the body json is valid tho
	Conf := make(map[string]string)
	err = json.Unmarshal(data, &Conf)
	if nil != err {
		fmt.Print("> Config file content is not a valid json")
		os.Exit(0)
		return
	}

	// finally we write all given configs into our config Data map ### need to change this rn we can only have required configs Oo what the fuck was i tinking
	for _, name := range requiredConfigs {
		value, ok := Conf[name]
		if ok {
			Data[name] = value
		}
	}
}
