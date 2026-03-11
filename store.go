package main

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type Item struct {
	item      string
	count     int
	firstSeen time.Time
	visits    []time.Time
	starred   bool
}

type store struct {
	dir string
}

func newStore(dir string) *store {
	return &store{
		dir: dir,
	}
}

func (s *store) listNs() (ns []string, err error) {
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

func (s *store) loadItems(ns string) (items []Item, err error) {
	f, err := s.ensureNs(ns)
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

func (s *store) saveItems(ns string, items ...Item) error {
	f, err := s.ensureNs(ns)
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
		itemSet[it.item] = struct{}{}
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

func (s *store) getItem(ns, name string) (*Item, error) {
	f, err := s.ensureNs(ns)
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
		if it.item == name {
			return &it, nil
		}
	}
	return nil, scanner.Err()
}

func (s *store) ensureNs(ns string) (*os.File, error) {
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

func trimVisits(visits []time.Time, max int) []time.Time {
	if len(visits) > max {
		return visits[len(visits)-max:] // keep most recent
	}
	return visits
}

func parseItem(line string) (Item, error) {
	parts := strings.Split(line, "\t")
	if len(parts) < 5 {
		return Item{}, fmt.Errorf("bad columns: %d", len(parts))
	}
	it := Item{}

	it.item = parts[0]

	cnt, err := strconv.Atoi(parts[1])
	if err != nil {
		return Item{}, fmt.Errorf("invalid int count: %v, err: %v", parts[1], err)
	}
	it.count = cnt

	ts, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		return Item{}, fmt.Errorf("invalid int64 firstSeen: %v, err: %v", parts[2], err)
	}
	it.firstSeen = time.Unix(ts, 0)

	vis := strings.SplitSeq(parts[3], ";")
	for v := range vis {
		ts, err = strconv.ParseInt(v, 10, 64)
		if err != nil {
			slog.Warn("invalid int64", "visits", v, "err", err)
			continue
		}
		it.visits = append(it.visits, time.Unix(ts, 0))
	}
	it.visits = trimVisits(it.visits, 10)

	it.starred = parts[4] == "y"
	return it, nil
}

func writeItems(w *bufio.Writer, items ...Item) {
	for _, it := range items {
		it.visits = trimVisits(it.visits, 10)

		parts := [5]string{}
		parts[0] = it.item
		parts[1] = strconv.Itoa(it.count)
		parts[2] = strconv.FormatInt(it.firstSeen.Unix(), 10)

		var visits []string
		for _, v := range it.visits {
			visits = append(visits, strconv.FormatInt(v.Unix(), 10))
		}
		parts[3] = strings.Join(visits, ";")
		parts[4] = "n"
		if it.starred {
			parts[4] = "y"
		}
		w.WriteString(strings.Join(parts[:], "\t") + "\n")
	}
}
