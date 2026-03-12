package cmd

import "github.com/shv-ng/frec/internal/store"

func StarItem(ns, name string, star bool) error {
	s, err := store.New()
	if err != nil {
		return err
	}
	item, err := s.GetItem(ns, name)
	if err != nil {
		return err
	}
	item.Starred = star
	return s.UpsertItems(ns, item)
}
