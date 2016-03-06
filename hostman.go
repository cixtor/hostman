package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"regexp"
	"strings"
)

// Hostman is a library with methods that allow the interaction with the Unix
// /etc/hosts file. Operations such as insertion, deletion, update, aliasing and
// export are implemented with public methods that can be accessed by 3rd-party
// libraries.
type Hostman struct{}

// Entries is a list of objects with attributes representing the information
// contained on each valid line found in the /etc/hosts file. The list of
// attributes includes the IP address, hostname, optional aliases, whether the
// line is commented or not and the raw string before the formalization.
type Entries []Entry

// Entry contains information of each host entry.
type Entry struct {
	Address  string
	Domain   string
	Aliases  []string
	Disabled bool
	Raw      string
}

func (h *Hostman) Config() string {
	_, err := os.Stat(*config)

	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}

	return *config
}

func (h *Hostman) Save(entries Entries) error {
	var final string

	for _, entry := range entries {
		if entry.Disabled {
			final += "#"
		}

		final += entry.Raw + "\n"
	}

	return ioutil.WriteFile(h.Config(), []byte(final), 0644)
}

func (h *Hostman) ParseEntry(line string) (Entry, error) {
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

func (h *Hostman) Entries() (Entries, error) {
	file, err := os.Open(h.Config())

	if err != nil {
		return nil, err
	}

	defer file.Close()

	var entries Entries
	var entry Entry
	var line string

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line = scanner.Text()
		entry, err = h.ParseEntry(line)

		if err == nil {
			entries = append(entries, entry)
		}
	}

	return entries, nil
}

func (h *Hostman) PrintEntries(entries Entries) {
	result, err := json.MarshalIndent(entries, "", "\x20\x20")

	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}

	fmt.Printf("%s\n", result)
}

func (h *Hostman) AlreadyExists(entry Entry) bool {
	entries := h.Entries()

	for _, current := range entries {
		if current.Raw == entry.Raw {
			return true
		}
	}

	return false
}

func (h *Hostman) InArray(haystack []string, needle string) bool {
	for _, value := range haystack {
		if value == needle {
			return true
		}
	}

	return false
}

func (h *Hostman) RawLines(entries Entries) []string {
	var lines []string

	for _, entry := range entries {
		lines = append(lines, entry.Raw)
	}

	return lines
}

func (h *Hostman) RemoveEntryAlias(entry Entry, alias string) Entry {
	var refactored []string

	for _, dalias := range entry.Aliases {
		if dalias != alias {
			refactored = append(refactored, dalias)
		}
	}

	entry.Aliases = refactored

	return entry
}

func (h *Hostman) EnableOrDisableEntries(entries Entries, action string) {
	current := h.Entries()

	var refactored Entries
	var lines []string = h.RawLines(entries)

	for _, entry := range current {
		if h.InArray(lines, entry.Raw) {
			fmt.Println(entry.Raw)

			if action != "remove" {
				entry.Disabled = (action == "disable")
				refactored = append(refactored, entry)
			}
		} else {
			refactored = append(refactored, entry)
		}
	}

	h.Save(refactored)
}

func (h *Hostman) RemoveEntries(entries Entries) {
	h.EnableOrDisableEntries(entries, "remove")
}

func (h *Hostman) EnableEntries(entries Entries) {
	h.EnableOrDisableEntries(entries, "enable")
}

func (h *Hostman) DisableEntries(entries Entries) {
	h.EnableOrDisableEntries(entries, "disable")
}

func (h *Hostman) AddEntry(line string) error {
	re := regexp.MustCompile(`^([0-9a-f:\.]{7,39})@(\S+)$`)
	parts := re.FindStringSubmatch(line)

	if len(parts) < 3 {
		return errors.New("Invalid host entry format")
	}

	line = strings.Replace(line, "@", "\t", 1)
	line = strings.Replace(line, ",", "\x20", -1)
	entry, err := h.ParseEntry(line)

	if err != nil {
		return err
	}

	if h.AlreadyExists(entry) {
		return errors.New("Entry is already in hosts file")
	}

	var config string = h.Config()

	file, err := os.OpenFile(config, os.O_APPEND|os.O_RDWR, 0644)

	if err != nil {
		return err
	}

	defer file.Close()

	_, err = io.WriteString(file, entry.Raw+"\n")

	return err
}

func (h *Hostman) ExportEntries() {
	entries := h.Entries()
	h.PrintEntries(entries)
}

func (h *Hostman) SearchEntry(query string) {
	var matches Entries

	entries := h.Entries()
	printResults := (!*export && !*enable && !*disable && !*remove)

	for _, entry := range entries {
		if strings.Contains(entry.Raw, query) {
			matches = append(matches, entry)

			if printResults == true {
				fmt.Printf("%s\n", entry.Raw)
			}
		}
	}

	if *export == true {
		h.PrintEntries(matches)
	} else if *enable == true {
		h.EnableEntries(matches)
	} else if *disable == true {
		h.DisableEntries(matches)
	} else if *remove == true {
		h.RemoveEntries(matches)
	}
}
