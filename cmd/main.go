package main

import (
	"flag"
	"fmt"
	"github.com/cixtor/hostman"
)

var add = flag.String("add", "", "Add new entry to the hosts file")
var config = flag.String("config", "/etc/hosts", "Absolute path of the hosts file")
var search = flag.String("search", "", "Search address or domain in the hosts file")
var disable = flag.Bool("disable", false, "Disable entries from the hosts file")
var enable = flag.Bool("enable", false, "Enable entries from the hosts file")
var remove = flag.Bool("remove", false, "Remove entries from the hosts file")
var export = flag.Bool("export", false, "List entries from the hosts file")

func jsonEncodeEntries(manager *hostman.Hostman, entries hostman.Entries) {
	out, err := manager.Export(entries)

	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}

	fmt.Printf("%s\n", out)
}

func main() {
	flag.Usage = func() {
		fmt.Println("Hostman (Hosts Manager)")
		fmt.Println("  http://cixtor.com/")
		fmt.Println("  https://github.com/cixtor/hostman")
		fmt.Println("  https://en.wikipedia.org/wiki/Hosts_(file)")
		fmt.Println("Usage:")
		flag.PrintDefaults()
		fmt.Println("Examples:")
		fmt.Println("  hostman -search example")
		fmt.Println("  hostman -search example -export")
		fmt.Println("  hostman -search example -remove")
		fmt.Println("  hostman -search 127.0.0.1 -enable")
		fmt.Println("  hostman -search 127.0.0.1 -disable")
		fmt.Println("  hostman -add 127.0.0.1@example.com")
		fmt.Println("  hostman -add 127.0.0.1@example.com,example.org")
		fmt.Println("  hostman -add 127.0.0.1@example.com,example.org,example.net")
		fmt.Println("  hostman -export (default: /etc/hosts)")
		fmt.Println("  hostman -config /tmp/hosts -export")
	}

	flag.Parse()

	manager, err := hostman.New(*config)

	if err != nil {
		fmt.Println(err)
		return
	}

	defer manager.Close()

	manager.Load()

	if *add != "" {
		if err := manager.Add(*add); err != nil {
			fmt.Println(err)
		}
		return
	}

	if *search != "" {
		result := manager.Search(*search)

		if *remove {
			if err := manager.Remove(result); err != nil {
				fmt.Println(err)
			}
			return
		}

		if *enable {
			if err := manager.Enable(result); err != nil {
				fmt.Println(err)
			}
			return
		}

		if *disable {
			if err := manager.Disable(result); err != nil {
				fmt.Println(err)
			}
			return
		}

		jsonEncodeEntries(manager, result)
		return
	}

	if *export {
		jsonEncodeEntries(manager, manager.Entries())
		return
	}
}
