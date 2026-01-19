package main

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"slices"
	"strings"
)

func setStandardPack() map[string]int {
	cardsInPack := map[string]int{
		"Ordinary":    11,
		"Exceptional": 3,
		"Elite":       0,
		"Unique":      0,
		"Avatar":	   0,
	}
	//20% of unique
	uniqueProbability, _ := rand.Int(rand.Reader, big.NewInt(int64(5)))
	if uniqueProbability.Int64() == 0 {
		cardsInPack["Unique"] += 1
	} else {
		cardsInPack["Elite"] += 1
	}
	return cardsInPack
}

func setPack(cmd command) (map[string]int, error) {
	cardsInPack := map[string]int{}

	if tag := slices.Index(cmd.arguments, "-p"); tag != -1 {
		if len(cmd.arguments) >= tag+2 && cmd.arguments[tag+1][0:1] != "-" {
			packType := strings.Title(cmd.arguments[tag+1])

			switch packType {
			case "Standard":
				cardsInPack = setStandardPack()
			case "Pauper":
				cardsInPack = map[string]int{
					"Ordinary":    11,
					"Exceptional": 4,
				}
			case "Pyramid":
				cardsInPack = map[string]int{
					"Ordinary":    8,
					"Exceptional": 4,
					"Elite":       2,
					"Unique":      1,
				}
			case "Ordinary", "O":
				cardsInPack = map[string]int{
					"Ordinary": 15,
				}
			case "Exceptional", "X":
				cardsInPack = map[string]int{
					"Exceptional": 15,
				}
			case "Elite", "E":
				cardsInPack = map[string]int{
					"Elite": 15,
				}
			case "Unique", "U":
				cardsInPack = map[string]int{
					"Unique": 15,
				}
			case "Custom":
				var o int
				var x int
				var e int
				var u int
				fmt.Println("Enter the number of cards (>= 0) of the following type you want to add to the pack:")
				fmt.Println("Ordinary:")
				fmt.Scanln(&o)
				fmt.Println("Exceptional:")
				fmt.Scanln(&x)
				fmt.Println("Elite:")
				fmt.Scanln(&e)
				fmt.Println("Unique:")
				fmt.Scanln(&u)
				fmt.Println("")
				cardsInPack = map[string]int{
					"Ordinary":    o,
					"Exceptional": x,
					"Elite":       e,
					"Unique":      u,
				}
			case "Random":
				var o int
				var x int
				var e int
				var u int
				leftInPack := 15

				randomOrdinary, _ := rand.Int(rand.Reader, big.NewInt(int64(leftInPack+1)))
				o = int(randomOrdinary.Int64())
				leftInPack -= o

				randomExceptional, _ := rand.Int(rand.Reader, big.NewInt(int64(leftInPack+1)))
				x = int(randomExceptional.Int64())
				leftInPack -= x

				randomElite, _ := rand.Int(rand.Reader, big.NewInt(int64(leftInPack+1)))
				e = int(randomElite.Int64())
				leftInPack -= e

				u = leftInPack

				cardsInPack = map[string]int{
					"Ordinary":    o,
					"Exceptional": x,
					"Elite":       e,
					"Unique":      u,
				}
			default:
				cardsInPack = setStandardPack()
			}

		} else {
			return nil, errors.New("No pack configuration was given after -p tag")
		}
	} else {
		cardsInPack = setStandardPack()
	}
	return cardsInPack, nil
}
