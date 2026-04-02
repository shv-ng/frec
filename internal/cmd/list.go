package cmd

import (
	"cmp"
	"fmt"
	"os"
	"slices"
	"strings"
	"text/tabwriter"

	"github.com/shv-ng/frec/internal/score"
	"github.com/shv-ng/frec/internal/store"
)

type scoredItem struct {
	item  *store.Item
	score int64
}

func ListItems(ns, format string, limit, since int, starred bool) error {
	s, err := store.New()
	if err != nil {
		return err
	}
	items, err := s.LoadItems(ns)
	if err != nil {
		return err
	}
	var scoredItems []scoredItem
	for _, it := range items {
		if starred && !it.Starred {
			continue
		}

		if since != -1 && score.DaysSince(it.LastSeen) > int64(since) {
			continue
		}

		si := scoredItem{}
		si.score = score.CalculateScore(it.FirstSeen, it.Visits, int64(it.Count), it.Starred)
		si.item = it
		scoredItems = append(scoredItems, si)
	}

	slices.SortFunc(scoredItems, func(a, b scoredItem) int {
		if n := cmp.Compare(b.score, a.score); n != 0 {
			return n
		}
		return cmp.Compare(b.item.LastSeen, a.item.LastSeen)
	})
	if limit != -1 && limit < len(scoredItems) {
		scoredItems = scoredItems[:limit]
	}
	printItem(scoredItems, unsanetizeFormat(format))
	return nil
}

func ListNs(format string) error {
	s, err := store.New()
	if err != nil {
		return err
	}
	ns, err := s.ListNs()
	if err != nil {
		return err
	}
	printNs(ns, unsanetizeFormat(format))
	return nil
}

func unsanetizeFormat(format string) string {
	format = strings.ReplaceAll(format, "\\n", "\n")
	format = strings.ReplaceAll(format, "\\t", "\t")
	return format
}

func printItem(items []scoredItem, format string) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
	defer w.Flush()
	for _, si := range items {
		star := ""
		if si.item.Starred {
			star = "*"
		}
		out := strings.ReplaceAll(format, "${name}", si.item.Name)
		out = strings.ReplaceAll(out, "${star}", star)
		out = strings.ReplaceAll(out, "${score}", fmt.Sprintf("%d", si.score))
		out = strings.ReplaceAll(out, "${count}", fmt.Sprintf("%d", si.item.Count))
		out = strings.ReplaceAll(out, "${firstseen}", fmt.Sprintf("%d", score.DaysSince(si.item.FirstSeen)))
		out = strings.ReplaceAll(out, "${lastseen}", fmt.Sprintf("%d", score.DaysSince(si.item.LastSeen)))

		fmt.Fprint(w, out)
	}
}

func printNs(val []string, format string) {
	for _, v := range val {
		out := strings.ReplaceAll(format, "${ns}", v)
		fmt.Print(out)
	}
}
