package main

import (
	"github.com/voodooEntity/gits"
	gitsTypes "github.com/voodooEntity/gits/src/types"
	"github.com/voodooEntity/gitsapi"
)

func main() {

	// first init gits storage
	gits.Init(gitsTypes.PersistenceConfig{
		RotationEntriesMax:           1000000,
		Active:                       true,
		PersistenceChannelBufferSize: 10000000,
	})

	// than start the http server
	gitsapi.Start()
}
