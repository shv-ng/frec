package store

import (
	"bufio"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"
)

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

	it.Item = parts[0]

	cnt, err := strconv.Atoi(parts[1])
	if err != nil {
		return Item{}, fmt.Errorf("invalid int count: %v, err: %v", parts[1], err)
	}
	it.Count = cnt

	ts, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		return Item{}, fmt.Errorf("invalid int64 firstSeen: %v, err: %v", parts[2], err)
	}
	it.FirstSeen = time.Unix(ts, 0)

	vis := strings.SplitSeq(parts[3], ";")
	for v := range vis {
		ts, err = strconv.ParseInt(v, 10, 64)
		if err != nil {
			slog.Warn("invalid int64", "visits", v, "err", err)
			continue
		}
		it.Visits = append(it.Visits, time.Unix(ts, 0))
	}
	it.Visits = trimVisits(it.Visits, 10)

	it.Starred = parts[4] == "y"
	return it, nil
}

func writeItems(w *bufio.Writer, items ...Item) {
	for _, it := range items {
		it.Visits = trimVisits(it.Visits, 10)

		parts := [5]string{}
		parts[0] = it.Item
		parts[1] = strconv.Itoa(it.Count)
		parts[2] = strconv.FormatInt(it.FirstSeen.Unix(), 10)

		var visits []string
		for _, v := range it.Visits {
			visits = append(visits, strconv.FormatInt(v.Unix(), 10))
		}
		parts[3] = strings.Join(visits, ";")
		parts[4] = "n"
		if it.Starred {
			parts[4] = "y"
		}
		w.WriteString(strings.Join(parts[:], "\t") + "\n")
	}
}
