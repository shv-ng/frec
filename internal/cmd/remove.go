package cmd

import "github.com/shv-ng/frec/internal/store"

func RemoveItem(ns, name string) error {
	s, err := store.New()
	if err != nil {
		return err
	}
	items, err := s.LoadItems(ns)
	if err != nil {
		return err
	}
	var tmp []*store.Item
	for _, it := range items {
		if it.Name != name {
			tmp = append(tmp, it)
		}
	}

	return s.ReplaceItems(ns, tmp...)
}

func RemoveNs(ns string) error {
	s, err := store.New()
	if err != nil {
		return err
	}
	return s.RemoveNs(ns)
}
