package main

import (
	"github.com/voodooEntity/archivist"
	"github.com/voodooEntity/gits"
	"github.com/voodooEntity/gitsapi"
	"github.com/voodooEntity/gitsapi/src/config"
)

func main() {
	config.Init(make(map[string]string))
	archivist.Init(config.GetValue("LOG_LEVEL"), config.GetValue("LOG_TARGET"), config.GetValue("LOG_PATH"))
	gits.NewInstance("api")
	gitsapi.Start()
}
