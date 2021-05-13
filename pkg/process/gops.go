// +build gops

package process

import (
	"log"

	"github.com/google/gops/agent"
)

// init initializes gops agent if `gops` tag is enabled
//
// it is in this package as this is one of the most central
// packages required by tests and deployable engines
func init() {
	if err := agent.Listen(agent.Options{
		ShutdownCleanup: true,
	}); err != nil {
		log.Fatal(err)
	}
}
