package store

import (
	"bufio"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

type Item struct {
	Name      string
	Count     int
	FirstSeen int64
	LastSeen  int64
	Visits    []int64
	Starred   bool
}

type store struct {
	dir string
}

func New(overridePath ...string) (*store, error) {
	var dir string
	var err error

	if len(overridePath) > 0 && overridePath[0] != "" {
		dir = overridePath[0]
	} else {
		dir, err = dataDir()
		if err != nil {
			return nil, err
		}
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	return &store{dir: dir}, nil
}

func (s *store) ListNs() (ns []string, err error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			name := entry.Name()
			ns = append(ns, strings.TrimSuffix(name, ".tsv"))
		}
	}
	sort.Strings(ns)
	return
}

func (s *store) LoadItems(ns string) (items []*Item, err error) {
	f, err := s.openNs(ns)
	if err != nil {
		return
	}
	defer unlockFile(f)
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		it, err := parseItem(scanner.Text())
		if err != nil {
			slog.Warn("skip row", "err", err)
			continue
		}
		items = append(items, it)
	}
	return
}

func (s *store) UpsertItems(ns string, items ...*Item) error {
	f, err := s.openNs(ns)
	if err != nil {
		return err
	}
	defer unlockFile(f)
	defer f.Close()
	tmp, err := os.CreateTemp(s.dir, ns+"*.tsv")
	if err != nil {
		return err
	}
	defer tmp.Close()

	scanner := bufio.NewScanner(f)
	writer := bufio.NewWriter(tmp)

	itemSet := make(map[string]struct{})
	for _, it := range items {
		itemSet[it.Name] = struct{}{}
	}

	for scanner.Scan() {
		parts := strings.Split(scanner.Text(), "\t")
		if _, ok := itemSet[parts[0]]; ok {
			continue
		}
		writer.WriteString(strings.Join(parts, "\t") + "\n")
	}
	writeItems(writer, items...)

	if err := scanner.Err(); err != nil {
		return err
	}
	if err := writer.Flush(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path.Join(s.dir, ns+".tsv"))
}

func (s *store) ReplaceItems(ns string, items ...*Item) error {
	f, err := s.openNs(ns)
	if err != nil {
		return err
	}
	defer unlockFile(f)
	defer f.Close()

	tmp, err := os.CreateTemp(s.dir, ns+"*.tsv")
	if err != nil {
		return err
	}
	defer tmp.Close()

	scanner := bufio.NewScanner(f)
	writer := bufio.NewWriter(tmp)

	itemSet := make(map[string]*Item)
	for _, it := range items {
		itemSet[it.Name] = it
	}

	for scanner.Scan() {
		parts := strings.Split(scanner.Text(), "\t")
		if _, ok := itemSet[parts[0]]; ok {
			writer.WriteString(strings.Join(parts, "\t") + "\n")
			delete(itemSet, parts[0])
		}
	}
	for _, it := range itemSet {
		writeItems(writer, it)
	}

	if err := scanner.Err(); err != nil {
		return err
	}
	if err := writer.Flush(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path.Join(s.dir, ns+".tsv"))
}

func (s *store) GetItem(ns, name string) (*Item, error) {
	f, err := s.openNs(ns)
	if err != nil {
		return nil, err
	}
	defer unlockFile(f)
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {

		it, err := parseItem(scanner.Text())
		if err != nil {
			return nil, err
		}
		if it.Name == name {
			return it, nil
		}
	}
	return nil, scanner.Err()
}

func (s *store) openNs(ns string) (*os.File, error) {
	f, err := os.OpenFile(path.Join(s.dir, ns+".tsv"), os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	if err := lockFile(f); err != nil {
		f.Close()
		return nil, err
	}
	return f, nil
}

func (s *store) RemoveNs(ns string) error {
	p := path.Join(s.dir, ns+".tsv")
	return os.Remove(p)
}

func dataDir() (string, error) {
	base := os.Getenv("XDG_DATA_HOME") // Linux
	if base == "" {
		base = os.Getenv("APPDATA") // Windows → C:\Users\<user>\AppData\Roaming
	}
	if base == "" {
		home, err := os.UserHomeDir() // fallback both
		if err != nil {
			return "", err
		}
		base = filepath.Join(home, ".local", "share")
	}
	return filepath.Join(base, "frec"), nil
}
