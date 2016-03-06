package main

import (
	"flag"
	"fmt"
)

var add = flag.String("add", "", "Add new entry to the hosts file")
var config = flag.String("config", "/etc/hosts", "Absolute path of the hosts file")
var search = flag.String("search", "", "Search address or domain in the hosts file")
var disable = flag.Bool("disable", false, "Disable entries from the hosts file")
var enable = flag.Bool("enable", false, "Enable entries from the hosts file")
var remove = flag.Bool("remove", false, "Remove entries from the hosts file")
var export = flag.Bool("export", false, "List entries from the hosts file")

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

	var manager Hostman

	if *add != "" {
		manager.AddEntry(*add)
		return
	}

	if *search != "" {
		manager.SearchEntry(*search)
		return
	}

	if *export {
		manager.ExportEntries()
		return
	}
}
