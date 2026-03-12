package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/shv-ng/frec/internal/store"
)

func SyncItems(ns string, null bool) error {
	items, err := readFromStdin(null)
	if err != nil {
		return err
	}

	if len(items) == 0 {
		return nil
	}

	s, err := store.New()
	if err != nil {
		return err
	}

	existings, err := s.LoadItems(ns)
	if err != nil {
		return err
	}

	es := make(map[string]*store.Item)
	for _, ex := range existings {
		es[ex.Name] = ex
	}

	now := time.Now().Unix()
	var tmp []*store.Item
	for _, name := range items {
		it, ok := es[name]
		if !ok {
			it = &store.Item{
				Name:      name,
				Count:     0,
				FirstSeen: now,
				LastSeen:  now,
				Visits:    []int64{now},
				Starred:   false,
			}
		}
		tmp = append(tmp, it)
	}

	return s.ReplaceItems(ns, tmp...)
}

func readFromStdin(null bool) (out []string, err error) {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return
	}

	if (stat.Mode() & os.ModeCharDevice) != 0 {
		fmt.Println("(Reading from stdin... Press Ctrl+D when finished)")
	}

	scanner := bufio.NewScanner(os.Stdin)
	if null {
		scanner.Split(nullSplitFunc)
	}
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		out = append(out, line)
	}
	err = scanner.Err()
	return

}

func nullSplitFunc(data []byte, atEOF bool) (advance int, token []byte, err error) {

	// Return nothing if at end of file and no data passed
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	if i := strings.Index(string(data), "\000"); i >= 0 {
		return i + 1, data[0:i], nil
	}

	// If at end of file with data return the data
	if atEOF {
		return len(data), data, nil
	}

	return
}
