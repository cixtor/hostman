package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
)

var add = flag.String("add", "", "Add new entry to the hosts file")
var config = flag.String("config", "/etc/hosts", "Absolute path of the hosts file")
var search = flag.String("search", "", "Search address or domain in the hosts file")
var disable = flag.Bool("disable", false, "Disable entries from the hosts file")
var remove = flag.Bool("remove", false, "Remove entries from the hosts file")
var export = flag.Bool("export", false, "List entries from the hosts file")

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

func (obj *Hostman) Save(entries Entries) {
	var final string

	for _, entry := range entries {
		if entry.Disabled {
			final += "#"
		}

		final += entry.Raw + "\n"
	}

	err := ioutil.WriteFile(obj.Config(), []byte(final), 0644)

	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}

	os.Exit(0)
}

func (obj *Hostman) ParseEntry(line string) (Entry, error) {
	var entry Entry
	var addresses []string
	var sections []string
	var quantity int

	line = strings.TrimSpace(line)

	if line == "" {
		return Entry{}, errors.New("Host entry is empty")
	}

	line = strings.Replace(line, "\x20", "\t", -1)
	sections = strings.Split(line, "\t")

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

	if entry.Address[0] == 0x23 {
		entry.Disabled = true
		entry.Address = entry.Address[1:len(entry.Address)]
	}

	if quantity > 2 {
		entry.Aliases = addresses[2:quantity]
	}

	entry.Raw = fmt.Sprintf("%s\t%s", entry.Address, entry.Domain)

	if len(entry.Aliases) > 0 {
		var daliases string = strings.Join(entry.Aliases, "\x20")
		entry.Raw += fmt.Sprintf("\x20%s", daliases)
	}

	return entry, nil
}

func (obj *Hostman) Entries() Entries {
	file, err := os.Open(obj.Config())

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

func (obj *Hostman) PrintEntries(entries Entries) {
	result, err := json.MarshalIndent(entries, "", "\x20\x20")

	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("%s\n", result)
	os.Exit(0)
}

func (obj *Hostman) AlreadyExists(entry Entry) bool {
	entries := obj.Entries()

	for _, current := range entries {
		if current.Raw == entry.Raw {
			return true
		}
	}

	return false
}

func (obj *Hostman) InArray(haystack []string, needle string) bool {
	for _, value := range haystack {
		if value == needle {
			return true
		}
	}

	return false
}

func (obj *Hostman) RawLines(entries Entries) []string {
	var lines []string

	for _, entry := range entries {
		lines = append(lines, entry.Raw)
	}

	return lines
}

func (obj *Hostman) RemoveEntryAlias(entry Entry, alias string) Entry {
	var refactored []string

	for _, dalias := range entry.Aliases {
		if dalias != alias {
			refactored = append(refactored, dalias)
		}
	}

	entry.Aliases = refactored

	return entry
}

func (obj *Hostman) RemoveEntries(entries Entries) {
	current := obj.Entries()
	var refactored Entries
	var lines []string = obj.RawLines(entries)

	for _, entry := range current {
		if obj.InArray(lines, entry.Raw) {
			fmt.Println(entry.Raw)
		} else {
			refactored = append(refactored, entry)
		}
	}

	obj.Save(refactored)
}

func (obj *Hostman) DisableEntries(entries Entries) {
	current := obj.Entries()
	var refactored Entries
	var lines []string = obj.RawLines(entries)

	for _, entry := range current {
		if obj.InArray(lines, entry.Raw) {
			entry.Disabled = true
			fmt.Println(entry.Raw)
		}

		refactored = append(refactored, entry)
	}

	obj.Save(refactored)
}

func (obj *Hostman) AddEntry(line string) {
	re := regexp.MustCompile(`^([0-9a-f:\.]{7,39})@(\S+)$`)
	var parts []string = re.FindStringSubmatch(line)

	if len(parts) < 3 {
		fmt.Println("Error: invalid host entry format")
		os.Exit(1)
	}

	line = strings.Replace(line, "@", "\t", 1)
	line = strings.Replace(line, ",", "\x20", -1)
	entry, err := obj.ParseEntry(line)

	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}

	if obj.AlreadyExists(entry) {
		fmt.Println("Error: entry is already in hosts file")
		os.Exit(1)
	}

	var config string = obj.Config()
	file, err := os.OpenFile(config, os.O_APPEND|os.O_RDWR, 0644)

	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}

	defer file.Close()
	_, err = io.WriteString(file, entry.Raw+"\n")

	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}

	os.Exit(0)
}

func (obj *Hostman) ExportEntries() {
	entries := obj.Entries()
	obj.PrintEntries(entries)
}

func (obj *Hostman) SearchEntry(query string) {
	var matches Entries
	entries := obj.Entries()
	var printResults bool = (!*export && !*disable && !*remove)

	for _, entry := range entries {
		if strings.Contains(entry.Raw, query) {
			matches = append(matches, entry)

			if printResults == true {
				fmt.Printf("%s\n", entry.Raw)
			}
		}
	}

	if *export == true {
		obj.PrintEntries(matches)
	} else if *disable == true {
		obj.DisableEntries(matches)
	} else if *remove == true {
		obj.RemoveEntries(matches)
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
		fmt.Println("Examples:")
		fmt.Println("  hostman -search example")
		fmt.Println("  hostman -search example -export")
		fmt.Println("  hostman -search example -remove")
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
	} else if *search != "" {
		manager.SearchEntry(*search)
	} else if *export == true {
		manager.ExportEntries()
	}

	flag.Usage()
	os.Exit(2)
}
