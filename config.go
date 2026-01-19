package main

import (
	"context"
)

func (s *state) updateConfig() error {
	//sets
	oldSets, err := s.database.GetSets(context.Background())
	if err != nil {
		return err
	}

	for _, st := range oldSets {
		s.config.Sets = append(s.config.Sets, st.Name)
	}

	s.config.Types = []string{"Avatar", "Minion", "Magic", "Aura", "Artifact", "Site"}
	s.config.Rarities = []string{"Ordinary", "Exceptional", "Elite", "Unique"}
	s.config.ALSirs = []string{"Dame Britomart", "Sir Agravaine", "Sir Balin", "Sir Bedivere", "Sir Bors the Younger", "Sir Gaheris", "Sir Gawain", "Sir Ironside", "Sir Kay", "Sir Lamorak", "Sir Morien", "Sir Pelleas", "Sir Perceval", "Sir Priamus", "Sir Tom Thumb", "Sir Tristan"}
	s.config.MiniSets = []string{"Dragonlord", "Promotional"}
	s.config.Exceptions = []string{"Relentless Crowd", "Winter River", "Erik's Curiosa", "Spellslinger", "Spire", "Stream", "Valley", "Wasteland"}
	//cards to be excluded:
	//Alpha: Relentless Crowd, Winter River, Erik's Curiosa
	//Beta: Spellslinger, Spire, Stream, Valley, Wasteland

	return nil
}
