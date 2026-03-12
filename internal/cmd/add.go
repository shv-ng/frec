package cmd

import (
	"time"

	"github.com/shv-ng/frec/internal/store"
)

func AddItem(ns, name string) error {
	s, err := store.New()
	if err != nil {
		return err
	}
	it, err := s.GetItem(ns, name)
	if err != nil {
		return err
	}
	now := time.Now().Unix()
	if it == nil {
		it = &store.Item{}
		it.Name = name
		it.FirstSeen = now
	}
	it.Count++
	it.LastSeen = now
	it.Visits = append(it.Visits, now)

	s.UpsertItems(ns, it)
	return nil
}
