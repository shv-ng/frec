package store

import (
	"bufio"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"
)

func trimVisits(visits []int64, max int) []int64 {
	if len(visits) > max {
		return visits[len(visits)-max:] // keep most recent
	}
	return visits
}

func parseItem(line string) (*Item, error) {
	parts := strings.Split(line, "\t")
	if len(parts) < 6 {
		return nil, fmt.Errorf("bad columns: %d", len(parts))
	}
	it := &Item{}

	it.Name = parts[0]

	cnt, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid int count: %v, err: %v", parts[1], err)
	}
	it.Count = cnt

	ts, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid int64 firstSeen: %v, err: %v", parts[2], err)
	}
	it.FirstSeen = ts

	ts, err = strconv.ParseInt(parts[3], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid int64 lastSeen: %v, err: %v", parts[3], err)
	}
	it.LastSeen = ts

	vis := strings.SplitSeq(parts[4], ";")
	for v := range vis {
		ts, err = strconv.ParseInt(v, 10, 64)
		if err != nil {
			slog.Warn("invalid int64", "visits", v, "err", err)
			continue
		}
		it.Visits = append(it.Visits, ts)
	}
	it.Visits = trimVisits(it.Visits, 10)

	it.Starred = parts[5] == "y"
	return it, nil
}

func writeItems(w *bufio.Writer, items ...*Item) {
	now := time.Now().Unix()
	for _, it := range items {
		it.Visits = trimVisits(it.Visits, 10)
		it.LastSeen = now

		parts := [6]string{}
		parts[0] = it.Name
		parts[1] = strconv.Itoa(it.Count)
		parts[2] = strconv.FormatInt(it.FirstSeen, 10)
		parts[3] = strconv.FormatInt(it.LastSeen, 10)

		var visits []string
		for _, v := range it.Visits {
			visits = append(visits, strconv.FormatInt(v, 10))
		}
		parts[4] = strings.Join(visits, ";")
		parts[5] = "n"
		if it.Starred {
			parts[5] = "y"
		}
		w.WriteString(strings.Join(parts[:], "\t") + "\n")
	}
}
