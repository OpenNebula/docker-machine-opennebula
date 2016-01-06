package main

import (
	"github.com/docker/machine/libmachine/drivers/plugin"
	"github.com/km4rcus/docker-machine-opennebula"
)

func main() {
	plugin.RegisterDriver(opennebula.NewDriver("", ""))
}
