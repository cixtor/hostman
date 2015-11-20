package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

var add = flag.String("add", "", "Add new entry to the hosts file")
var remove = flag.String("remove", "", "Remove entries from the hosts file")
var export = flag.Bool("export", false, "List entries from the hosts file")
var search = flag.String("search", "", "Search address or domain in the hosts file")
var config = flag.String("config", "/etc/hosts", "Absolute path of the hosts file")

type Hostman struct{}

type Entries []Entry

type Entry struct {
	Address  string
	Domain   string
	Aliases  []string
	Disabled bool
	Raw      string
}

func (obj *Hostman) Config() string {
	_, err := os.Stat(*config)

	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}

	return *config
}

func (obj *Hostman) ParseEntry(raw string) (Entry, error) {
	var entry Entry
	var addresses []string
	var sections []string
	var quantity int

	raw = strings.TrimSpace(raw)

	if raw == "" {
		return Entry{}, errors.New("Host entry is empty")
	}

	raw = strings.Replace(raw, "\x20", "\t", -1)
	sections = strings.Split(raw, "\t")

	for _, section := range sections {
		if section != "" {
			addresses = append(addresses, section)
		}
	}

	quantity = len(addresses)

	if quantity < 2 {
		return Entry{}, errors.New("Address and domain are required")
	}

	entry.Address = addresses[0]
	entry.Domain = addresses[1]
	entry.Disabled = entry.Address[0] == 0x23
	entry.Raw = strings.Join(addresses, "\x20")

	if quantity > 2 {
		entry.Aliases = addresses[2:quantity]
	}

	return entry, nil
}

func (obj *Hostman) Entries() Entries {
	config := obj.Config()
	file, err := os.Open(config)

	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}

	defer file.Close()

	var entries Entries
	var entry Entry
	var line string

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line = scanner.Text()
		entry, err = obj.ParseEntry(line)

		if err == nil {
			entries = append(entries, entry)
		}
	}

	return entries
}

func (obj *Hostman) AddEntry(entry string) {
	re := regexp.MustCompile(`^([0-9a-f:\.]{7,39})@(\S+)$`)
	var parts []string = re.FindStringSubmatch(entry)

	if len(parts) == 3 {
		var formatted string
		var address string = parts[1]
		var domains string = parts[2]
		var config string = obj.Config()
		file, err := os.OpenFile(config, os.O_APPEND|os.O_RDWR, 0644)

		if err != nil {
			fmt.Printf("Error: %s\n", err)
			os.Exit(1)
		}

		defer file.Close()
		domains = strings.Replace(domains, ",", "\x20", -1)
		formatted = fmt.Sprintf("%s\t%s\n", address, domains)
		_, err = io.WriteString(file, formatted)

		if err != nil {
			fmt.Printf("Error: %s\n", err)
			os.Exit(1)
		}

		os.Exit(0)
	} else {
		fmt.Println("Invalid format in host entry")
	}
}

func (obj *Hostman) PrintExportEntries(entries Entries) {
	result, err := json.MarshalIndent(entries, "", "\x20\x20")

	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("%s\n", result)
	os.Exit(0)
}

func (obj *Hostman) ExportEntries() {
	entries := obj.Entries()
	obj.PrintExportEntries(entries)
}

func (obj *Hostman) SearchEntry(query string) {
	entries := obj.Entries()
	var matches Entries

	for _, entry := range entries {
		if strings.Contains(entry.Raw, query) {
			matches = append(matches, entry)

			if *export == false {
				fmt.Printf("%s\n", entry.Raw)
			}
		}
	}

	if *export == true {
		obj.PrintExportEntries(matches)
	}

	os.Exit(0)
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

	if *add != "" {
		manager.AddEntry(*add)
	} else if *search != "" {
		manager.SearchEntry(*search)
	} else if *export == true {
		manager.ExportEntries()
	}

	flag.Usage()
	os.Exit(2)
}
