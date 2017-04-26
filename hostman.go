package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
)

// Hostman is a library with methods that allow the interaction with the Unix
// /etc/hosts file. Operations such as insertion, deletion, update, aliasing and
// export are implemented with public methods that can be accessed by 3rd-party
// libraries.
type Hostman struct {
	registry *os.File
	entries  Entries
	filename string
	count    int
}

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

// New creates a new instance of Hostman.
func New(filename string) (*Hostman, error) {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_RDWR|os.O_RDONLY, 0644)

	if err != nil {
		return nil, err
	}

	return &Hostman{registry: file, filename: filename}, nil
}

func (h *Hostman) inarray(haystack []string, needle string) bool {
	for _, value := range haystack {
		if value == needle {
			return true
		}
	}

	return false
}

// Close frees the file resource.
func (h *Hostman) Close() error {
	return h.registry.Close()
}

// Load reads the hosts file and parses its current content.
func (h *Hostman) Load() {
	scanner := bufio.NewScanner(h.registry)

	for scanner.Scan() {
		entry, err := h.Parse(scanner.Text())

		if err != nil {
			continue
		}

		h.entries = append(h.entries, entry)
		h.count++
	}
}

// Write inserts a new entry into the hosts file.
func (h *Hostman) Write() error {
	var final string

	for _, entry := range h.entries {
		if entry.Disabled {
			final += "#"
		}

		final += entry.Raw + "\n"
	}

	return ioutil.WriteFile(h.filename, []byte(final), 0644)
}

// Parse takes a string a tries to extract its associated attributes. If the
// string is not a valid hosts file entry then it returns a nil object with an
// error message explaining the failure.
func (h *Hostman) Parse(line string) (Entry, error) {
	var entry Entry
	var quantity int
	var sections []string
	var addresses []string

	line = strings.TrimSpace(line)

	if line == "" {
		return Entry{}, errors.New("Host entry is empty")
	}

	if len(line) >= 2 && line[0:2] == "#\x20" {
		return Entry{}, errors.New("Superfluous comment")
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

	if entry.Address[0] == '#' {
		entry.Disabled = true
		entry.Address = entry.Address[1:len(entry.Address)]
	}

	if quantity > 2 {
		entry.Aliases = addresses[2:quantity]
	}

	entry.Raw = entry.Address + "\t" + entry.Domain

	if len(entry.Aliases) > 0 {
		entry.Raw += "\x20" + strings.Join(entry.Aliases, "\x20")
	}

	return entry, nil
}

// Entries returns a list of objects representing the content of the hosts file.
// Invalid lines are ignored and no error messages will be passed back to the
// function caller as a result of a failure, you must treat the output always as
// a success no matter if the list of entries is empty.
func (h *Hostman) Entries() Entries {
	return h.entries
}

// AlreadyExists checks if an entry already exists in the hosts file.
func (h *Hostman) AlreadyExists(entry Entry) bool {
	for _, current := range h.entries {
		if current.Raw == entry.Raw {
			return true
		}
	}

	return false
}

// RawLines returns a list of strings representing the real content of the hosts
// file without transformations nor parsing nor auto-fixes. Notice that empty
// lines are ignored as well as comments except for commented entries.
func (h *Hostman) RawLines(entries Entries) []string {
	var lines []string

	for _, entry := range entries {
		lines = append(lines, entry.Raw)
	}

	return lines
}

func (h *Hostman) enableOrDisableEntries(entries Entries, action string) error {
	var refactored Entries

	lines := h.RawLines(entries)

	for _, entry := range h.entries {
		if h.inarray(lines, entry.Raw) {
			if action != "remove" {
				entry.Disabled = (action == "disable")
				refactored = append(refactored, entry)
			}
		} else {
			refactored = append(refactored, entry)
		}
	}

	h.entries = refactored

	return h.Write()
}

// Remove deletes an entry from the hosts file.
func (h *Hostman) Remove(entries Entries) error {
	return h.enableOrDisableEntries(entries, "remove")
}

// Enable uncomments an entry from the hosts file.
func (h *Hostman) Enable(entries Entries) error {
	return h.enableOrDisableEntries(entries, "enable")
}

// Disable comments an entry from the hosts file.
func (h *Hostman) Disable(entries Entries) error {
	return h.enableOrDisableEntries(entries, "disable")
}

// Add inserts an entry into the hosts file.
func (h *Hostman) Add(line string) error {
	re := regexp.MustCompile(`^([0-9a-f:\.]{7,39})@(\S+)$`)
	parts := re.FindStringSubmatch(line)

	if len(parts) < 3 {
		return errors.New("Invalid host entry format")
	}

	line = strings.Replace(line, "@", "\t", 1)
	line = strings.Replace(line, ",", "\x20", -1)

	entry, err := h.Parse(line)

	if err != nil {
		return err
	}

	if h.AlreadyExists(entry) {
		return errors.New("Entry is already in hosts file")
	}

	h.entries = append(h.entries, entry)

	return h.Write()
}

// RemoveAlias pops a domain alias associated to a host entry.
func (h *Hostman) RemoveAlias(entry Entry, alias string) Entry {
	var refactored []string

	for _, dalias := range entry.Aliases {
		if dalias != alias {
			refactored = append(refactored, dalias)
		}
	}

	entry.Aliases = refactored

	return entry
}

// Export JSON-encodes the entire hosts file and returns.
func (h *Hostman) Export(entries Entries) ([]byte, error) {
	return json.MarshalIndent(entries, "", "\x20\x20")
}

// Search locates and returns an entry from the hosts file.
func (h *Hostman) Search(query string) Entries {
	var matches Entries

	entries := h.Entries()

	for _, entry := range entries {
		if strings.Contains(entry.Raw, query) {
			matches = append(matches, entry)
		}
	}

	return matches
}
