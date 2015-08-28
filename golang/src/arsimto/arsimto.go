package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

func main() {
	collectUriPtr := flag.String("collect", "", "URI (typically user@host) to collect Asset data from")
	colourisedPtr := flag.Bool("c", false, "Should output be colourised")
        flag.Parse()

        if *collectUriPtr != "" {
            fmt.Println("Collecting from ", *collectUriPtr)
        } else {
            fmt.Println("Not collecting!")
        }

        if *colourisedPtr == true {
            fmt.Println("We will do colour")
        } else {
            fmt.Println("We will do plain black n white")
        }

	a := make(map[string]string)
	a["name"] = "mach100"
	a["foo"] = "bar"
	a["ip"] = "10.18.1.100"

	// convert this to JSON for output
	ja, err := json.Marshal(a)
	if err != nil {
		fmt.Println("error:", err)
	}

	fmt.Print("Object:")
	os.Stdout.Write(ja)
	fmt.Println()
}
