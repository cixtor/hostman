package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
)

const config string = "/etc/hosts"

var export = flag.Bool("export", false, "List entries from the hosts file")

type Hostman struct{}

type Entries []Entry

type Entry struct {
	Address  string
	Domain   string
	Aliases  []string
	Disabled bool
}

func (obj *Hostman) Entries() Entries {
	file, err := os.Open(config)

	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}

	defer file.Close()

	var line string
	var quantity int
	var sections []string
	var addresses []string
	var entries Entries
	var entry Entry

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line = scanner.Text()

		if line != "" {
			addresses = []string{}
			line = strings.Replace(line, "\x20", "\t", -1)
			sections = strings.Split(line, "\t")

			for _, section := range sections {
				if section != "" {
					addresses = append(addresses, section)
				}
			}

			quantity = len(addresses)

			if quantity >= 2 {
				entry.Address = addresses[0]
				entry.Domain = addresses[1]
				entry.Disabled = entry.Address[0] == 0x23

				if quantity > 2 {
					entry.Aliases = addresses[2:quantity]
				}

				entries = append(entries, entry)
			}
		}
	}

	return entries
}

func main() {
	flag.Usage = func() {
		fmt.Println("Hostman (Hosts Manager)")
		fmt.Println("  http://cixtor.com/")
		fmt.Println("  https://github.com/cixtor/hostman")
		fmt.Println("  https://en.wikipedia.org/wiki/Hosts_(file)")
		fmt.Println("Usage:")
		flag.PrintDefaults()
	}

	flag.Parse()

	var manager Hostman

	if *export == true {
		entries := manager.Entries()
		result, err := json.MarshalIndent(entries, "", "\x20\x20")

		if err != nil {
			fmt.Printf("Error: %s\n", err)
			os.Exit(1)
		}

		fmt.Printf("%s\n", result)
		os.Exit(0)
	}

	flag.Usage()
}
