package store

import (
	"bufio"
	"log/slog"
	"os"
	"path"
	"sort"
	"strings"
	"syscall"
	"time"
)

type Item struct {
	Item      string
	Count     int
	FirstSeen time.Time
	Visits    []time.Time
	Starred   bool
}

type store struct {
	dir string
}

func New(dir string) *store {
	return &store{
		dir: dir,
	}
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

func (s *store) LoadItems(ns string) (items []Item, err error) {
	f, err := s.EnsureNs(ns)
	if err != nil {
		return
	}
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

func (s *store) DumpItems(ns string, items ...Item) error {
	f, err := s.EnsureNs(ns)
	if err != nil {
		return err
	}
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
		itemSet[it.Item] = struct{}{}
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
	f.Close()
	return os.Rename(tmp.Name(), path.Join(s.dir, ns+".tsv"))
}

func (s *store) GetItem(ns, name string) (*Item, error) {
	f, err := s.EnsureNs(ns)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {

		it, err := parseItem(scanner.Text())
		if err != nil {
			return nil, err
		}
		if it.Item == name {
			return &it, nil
		}
	}
	return nil, scanner.Err()
}

func (s *store) EnsureNs(ns string) (*os.File, error) {
	f, err := os.OpenFile(path.Join(s.dir, ns+".tsv"), os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		f.Close()
		return nil, err
	}
	return f, nil
}
