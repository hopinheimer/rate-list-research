package flag

import "github.com/urfave/cli/v2"

var (
	Bootnode = &cli.StringFlag{
		Name:  "bootnode",
		Usage: "Initiate a bootnode with this",
		Value: "",
	}
	Create = &cli.StringFlag{
		Name:  "create",
		Usage: "Create a new bootnode for the network, this will save a new bootnode file in identities",
		Value: "false",
	}
)
