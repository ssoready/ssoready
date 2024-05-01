package main

import (
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/ssoready/prettyuuid"
)

func main() {
	format := prettyuuid.MustNewFormat("", "0123456789abcdefghijklmnopqrstuvwxyz")
	fmt.Println(format.Format(uuid.MustParse(os.Args[1])))
}
