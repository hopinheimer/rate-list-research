package main

import (
	"context"
	"log"
	"os"
	runtimeDebug "runtime/debug"
	"strconv"

	"github.com/chirag-parmar/rated-list/cli/flag"
	"github.com/chirag-parmar/rated-list/node"
	"github.com/chirag-parmar/rated-list/util"
	"github.com/urfave/cli/v2"
)

var appFlags = []cli.Flag{
	flag.Bootnode,
	flag.Create,
}

func main() {

	rctx, _ := context.WithCancel(context.Background())
	app := cli.App{
		Name:  "rated-list-dht",
		Usage: "This an experimental implementation for Rate List DHT",
		Action: func(ctx *cli.Context) error {
			if err := startNode(&rctx, ctx); err != nil {
				log.Fatal(err.Error())
				return err
			}
			return nil
		},
		Flags: appFlags,
	}

	defer func() {
		if x := recover(); x != nil {
			log.Fatal("Runtime panic: ", x, string(runtimeDebug.Stack()))
			panic(x)
		}
	}()

	if err := app.RunContext(rctx, os.Args); err != nil {
		log.Fatal(err.Error())
	}

}

func startNode(rctx *context.Context, c *cli.Context) error {

	util.LogHello()

	create, err := strconv.ParseBool(c.String(flag.Create.Name))
	if err != nil {
		log.Fatal("flag parse error", err)
		return err
	}

	rtnode, err := node.New(*rctx, c.String(flag.Bootnode.Name), create)
	if err != nil {
		log.Fatal("node creation failure", err)
		return err
	}

	log.Default().Println("starting ratedlist node ..... ")
	rtnode.Start()

	return nil
}
