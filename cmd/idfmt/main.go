package main

import (
	"flag"
	"fmt"

	"github.com/google/uuid"
	"github.com/ssoready/prettyuuid"
)

func main() {
	parse := flag.Bool("parse", false, "parse prettyuuid, output uuid")
	flag.Parse()
	format := prettyuuid.MustNewFormat("", "0123456789abcdefghijklmnopqrstuvwxyz")

	if *parse {
		id, err := format.Parse(flag.Arg(0))
		if err != nil {
			panic(err)
		}
		fmt.Println(uuid.UUID(id))
	} else {
		fmt.Println(format.Format(uuid.MustParse(flag.Arg(0))))
	}
}
