package main

import (
	"github.com/netscrn/homm3utils/defparse"
	"os"
)

func main() {
	err := defparse.ExtractDef(os.Args[1], os.Args[2])
	if err != nil {
		panic(err)
	}
}