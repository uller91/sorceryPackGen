package main

import (
	"slices"
)

func setAvatar(cardsInPack map[string]int, cmd command) (map[string]int, error) {
	if tag := slices.Index(cmd.arguments, "-a"); tag != -1 {
		cardsInPack["Ordinary"] -= 1
		cardsInPack["Avatar"] += 1
	}

	return cardsInPack, nil
}