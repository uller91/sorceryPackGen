package main

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"slices"
	"strings"
)

type Card struct {
	Name   string
	Rarity string
	Type   string
	Sets   string
}

// func getRandomCardsFromCollection[T any](s *state, collection []T, quantity int) []T
func getRandomCardsFromCollection(s *state, collection []Card, quantity int) []Card {
	if quantity >= len(collection) {
		fmt.Println("The collection is too small to get random cards. Returning full collection")
		return collection
	}

	if quantity < 0 {
		fmt.Println("The quantity of cards to return is less then 0. Returning 0 cards...")
		fmt.Println("")
		quantity = 0
	}

	randomCollection := []Card{}

	for i := 0; i < quantity; i++ {
		randomCardNumber, _ := rand.Int(rand.Reader, big.NewInt(int64(len(collection))))
		cardNumber := int(randomCardNumber.Int64())
		randomCard := collection[cardNumber]
		collection = slices.Delete(collection, cardNumber, cardNumber+1)

		//checking for cards that shouldn't be drawn
		if slices.Contains(s.config.Exceptions, randomCard.Name) {
			i -= 1
			continue
		}

		randomCollection = append(randomCollection, randomCard)
	}

	return randomCollection
}

func getRandomCardsFromCollectionByRarity(s *state, collection []Card, cardsQuantity, packsQuantity, rarityQuantity int, drafted map[string]int) []Card {
	if cardsQuantity >= len(collection) {
		fmt.Println("The collection is too small to get random cards. Returning full collection")
		return collection
	}

	if cardsQuantity < 0 {
		fmt.Println("The quantity of cards to return is less then 0. Returning 0 cards...")
		fmt.Println("")
		cardsQuantity = 0
	}

	randomCollection := []Card{}
	thisPack := []string{}

	for i := 0; i < cardsQuantity; i++ {
		randomCardNumber, _ := rand.Int(rand.Reader, big.NewInt(int64(len(collection))))
		cardNumber := int(randomCardNumber.Int64())
		randomCard := collection[cardNumber]

		//checking for cards that shouldn't be drawn
		if slices.Contains(s.config.Exceptions, randomCard.Name) || slices.Contains(thisPack, randomCard.Name) {
			//fmt.Println(randomCard.Name)
			i -= 1
			continue
		}

		//checking the collection if too many of such card is drawn
		quantityAlreadyDrafted, ok := drafted[randomCard.Name]
		if ok {
			if quantityAlreadyDrafted < rarityQuantity {
				drafted[randomCard.Name] += 1
			} else {
				i -= 1
				continue
			}
		} else {
			drafted[randomCard.Name] = 1
		}

		randomCollection = append(randomCollection, randomCard)
		thisPack = append(thisPack, randomCard.Name)
	}

	return randomCollection
}

/*
func generateOnePack(s *state, setName string, cardsInPack map[string]int) error {
	set, err := s.database.GetCardsBySet(context.Background(), setName)
	if err != nil {
		return err
	}

	cardsOrdinary := []Card{}
	cardsExceptional := []Card{}
	cardsElite := []Card{}
	cardsUnique := []Card{}
	cardsAvatar := []Card{}

	for _, setCard := range set {
		card, err := s.database.GetCard(context.Background(), setCard.CardID)
		if err != nil {
			return err
		}

		rarity := card.Rarity

		cardClean := Card{
			Name:   card.Name,
			Rarity: rarity,
			Type:   card.Type,
		}

		switch rarity {
		case "Ordinary":
			cardsOrdinary = append(cardsOrdinary, cardClean)
		case "Exceptional":
			cardsExceptional = append(cardsExceptional, cardClean)
		case "Elite":
			cardsElite = append(cardsElite, cardClean)
		case "Unique":
			if setName == "Arthurian Legends" && slices.Contains(s.config.ALSirs, card.Name) {
				cardsElite = append(cardsElite, cardClean)
				continue
			}
			cardsUnique = append(cardsUnique, cardClean)
		case "": //avatar
			cardsAvatar = append(cardsAvatar, cardClean)
		default:
			fmt.Println(cardClean.Rarity)
			return errors.New("Unknown rarity was found!")
		}
	}

	if slices.Contains(s.config.MiniSets, setName) {
		//redundant but will leave it here for now. Just in case
		fmt.Printf("Can't generate pack for %s mini-set! Use \"generate -s all\" to generate pack to include all the cards\n", setName)
	} else {
		pack := getRandomCardsFromCollection(s, cardsOrdinary, cardsInPack["Ordinary"])
		pack = append(pack, getRandomCardsFromCollection(s, cardsExceptional, cardsInPack["Exceptional"])...)
		pack = append(pack, getRandomCardsFromCollection(s, cardsElite, cardsInPack["Elite"])...)
		pack = append(pack, getRandomCardsFromCollection(s, cardsUnique, cardsInPack["Unique"])...)

		fmt.Printf("Random pack from %s set:\n", setName)
		fmt.Println("")

		for _, card := range pack {
			fmt.Printf("%-25v | %-10v | %-15v\n", card.Name, card.Type, card.Rarity)
		}
	}

	return nil
}
*/

func generateMiltiplePacks(s *state, setName string, cardsInPack map[string]int, packsQuantity int) error {
	set, err := s.database.GetCardsBySet(context.Background(), setName)
	if err != nil {
		return err
	}

	cardsOrdinary := []Card{}
	cardsExceptional := []Card{}
	cardsElite := []Card{}
	cardsUnique := []Card{}
	cardsAvatar := []Card{}

	for _, setCard := range set {
		card, err := s.database.GetCard(context.Background(), setCard.CardID)
		if err != nil {
			return err
		}

		rarity := card.Rarity

		cardClean := Card{
			Name:   card.Name,
			Rarity: rarity,
			Type:   card.Type,
		}

		switch rarity {
		case "Ordinary":
			cardsOrdinary = append(cardsOrdinary, cardClean)
		case "Exceptional":
			cardsExceptional = append(cardsExceptional, cardClean)
		case "Elite":
			cardsElite = append(cardsElite, cardClean)
		case "Unique":
			if setName == "Arthurian Legends" && slices.Contains(s.config.ALSirs, card.Name) {
				cardsElite = append(cardsElite, cardClean)
				continue
			}
			cardsUnique = append(cardsUnique, cardClean)
		case "": //avatar
			cardsAvatar = append(cardsAvatar, cardClean)
		default:
			return errors.New("Unknown rarity was found!")
		}
	}

	if len(cardsOrdinary)*4 < cardsInPack["Ordinary"]*packsQuantity || len(cardsExceptional)*3 < cardsInPack["Exceptional"]*packsQuantity || len(cardsElite)*2 < cardsInPack["Elite"]*packsQuantity || len(cardsUnique)*1 < cardsInPack["Unique"]*packsQuantity {
		return errors.New("You are trying to generate too many packs... or a single pack which is too big!")
	}

	if slices.Contains(s.config.MiniSets, setName) {
		//redundant but will leave it here for now. Just in case
		fmt.Printf("Can't generate pack for %s mini-set! Use \"generate -s all\" to generate pack to include all the cards\n", setName)
	} else {
		draftedOrdinary := map[string]int{}
		draftedExceptional := map[string]int{}
		draftedElite := map[string]int{}
		draftedUnique := map[string]int{}
		draftedAvatar := map[string]int{}

		quantityOrdinary := 4
		quantityExceptional := 3
		quantityElite := 2
		quantityUnique := 1
		quantityAvatar := 1

		packs := [][]Card{}

		for i := 0; i < packsQuantity; i++ {
			pack := getRandomCardsFromCollectionByRarity(s, cardsAvatar, cardsInPack["Avatar"], packsQuantity, quantityAvatar, draftedAvatar)
			pack = append(pack, getRandomCardsFromCollectionByRarity(s, cardsOrdinary, cardsInPack["Ordinary"], packsQuantity, quantityOrdinary, draftedOrdinary)...)
			pack = append(pack, getRandomCardsFromCollectionByRarity(s, cardsExceptional, cardsInPack["Exceptional"], packsQuantity, quantityExceptional, draftedExceptional)...)
			pack = append(pack, getRandomCardsFromCollectionByRarity(s, cardsElite, cardsInPack["Elite"], packsQuantity, quantityElite, draftedElite)...)
			pack = append(pack, getRandomCardsFromCollectionByRarity(s, cardsUnique, cardsInPack["Unique"], packsQuantity, quantityUnique, draftedUnique)...)

			packs = append(packs, pack)
		}

		if packsQuantity == 1 {
			fmt.Printf("Random pack from %s set:\n", setName)
			fmt.Println("")
		} else if packsQuantity > 1 {
			fmt.Printf("%v random packs from %s set:\n", packsQuantity, setName)
			fmt.Println("")
		} else {
			fmt.Println("0 packs was generated as requested...")
		}

		for _, pack := range packs {
			if packsQuantity > 1 {
				fmt.Println("Press the Enter to see the content of the next pack...")
				fmt.Scanln() // wait for Enter Key
			}

			for _, card := range pack {
				fmt.Printf("%-25v | %-10v | %-15v\n", card.Name, card.Type, card.Rarity)
			}
			fmt.Println("")
		}

	}

	return nil
}

func generateMultiplePacksAll(s *state, cardsInPack map[string]int, packsQuantity int) error {
	cards, err := s.database.GetAllCards(context.Background())
	if err != nil {
		return err
	}

	cardsOrdinary := []Card{}
	cardsExceptional := []Card{}
	cardsElite := []Card{}
	cardsUnique := []Card{}
	cardsAvatar := []Card{}

	for _, card := range cards {
		sets, err := s.database.GetSetsByCard(context.Background(), card.ID)
		if err != nil {
			return err
		}

		var setNames []string
		for _, set := range sets {
			setNames = append(setNames, set.Name)
		}

		rarity := card.Rarity
		setName := strings.Join(setNames, " / ")

		cardClean := Card{
			Name:   card.Name,
			Rarity: rarity,
			Type:   card.Type,
			Sets:   setName,
		}

		switch rarity {
		case "Ordinary":
			cardsOrdinary = append(cardsOrdinary, cardClean)
		case "Exceptional":
			cardsExceptional = append(cardsExceptional, cardClean)
		case "Elite":
			cardsElite = append(cardsElite, cardClean)
		case "Unique":
			if setName == "Arthurian Legends" && slices.Contains(s.config.ALSirs, card.Name) {
				cardsElite = append(cardsElite, cardClean)
				continue
			}
			cardsUnique = append(cardsUnique, cardClean)
		case "": //avatar
			cardsAvatar = append(cardsAvatar, cardClean)
		default:
			return errors.New("Unknown rarity was found!")
		}
	}

	draftedOrdinary := map[string]int{}
	draftedExceptional := map[string]int{}
	draftedElite := map[string]int{}
	draftedUnique := map[string]int{}
	draftedAvatar := map[string]int{}

	quantityOrdinary := 4
	quantityExceptional := 3
	quantityElite := 2
	quantityUnique := 1
	quantityAvatar := 1

	packs := [][]Card{}

	for i := 0; i < packsQuantity; i++ {
		pack := getRandomCardsFromCollectionByRarity(s, cardsAvatar, cardsInPack["Avatar"], packsQuantity, quantityAvatar, draftedAvatar)
		pack = append(pack, getRandomCardsFromCollectionByRarity(s, cardsOrdinary, cardsInPack["Ordinary"], packsQuantity, quantityOrdinary, draftedOrdinary)...)
		pack = append(pack, getRandomCardsFromCollectionByRarity(s, cardsExceptional, cardsInPack["Exceptional"], packsQuantity, quantityExceptional, draftedExceptional)...)
		pack = append(pack, getRandomCardsFromCollectionByRarity(s, cardsElite, cardsInPack["Elite"], packsQuantity, quantityElite, draftedElite)...)
		pack = append(pack, getRandomCardsFromCollectionByRarity(s, cardsUnique, cardsInPack["Unique"], packsQuantity, quantityUnique, draftedUnique)...)

		packs = append(packs, pack)
	}

	if packsQuantity == 1 {
		fmt.Println("Random pack from all sets:")
		fmt.Println("")
	} else if packsQuantity > 1 {
		fmt.Printf("%v random packs from all sets:\n", packsQuantity)
		fmt.Println("")
	} else {
		fmt.Println("0 packs was generated as requested...")
	}

	for _, pack := range packs {
		if packsQuantity > 1 {
			fmt.Println("Press the Enter to see the content of the next pack...")
			fmt.Scanln() // wait for Enter Key
		}

		for _, card := range pack {
			fmt.Printf("%-25v | %-10v | %-15v | %-10v\n", card.Name, card.Type, card.Rarity, card.Sets)
		}
		fmt.Println("")
	}

	return nil
}

/*
func generateOnePackAll(s *state, cardsInPack map[string]int) error {
	cards, err := s.database.GetAllCards(context.Background())
	if err != nil {
		return err
	}

	cardsOrdinary := []Card{}
	cardsExceptional := []Card{}
	cardsElite := []Card{}
	cardsUnique := []Card{}

	for _, card := range cards {
		sets, err := s.database.GetSetsByCard(context.Background(), card.ID)
		if err != nil {
			return err
		}

		var setNames []string
		for _, set := range sets {
			setNames = append(setNames, set.Name)
		}

		rarity := card.Rarity
		setName := strings.Join(setNames, " / ")

		cardClean := Card{
			Name:   card.Name,
			Rarity: rarity,
			Type:   card.Type,
			Sets:   setName,
		}

		switch rarity {
		case "Ordinary":
			cardsOrdinary = append(cardsOrdinary, cardClean)
		case "Exceptional":
			cardsExceptional = append(cardsExceptional, cardClean)
		case "Elite":
			cardsElite = append(cardsElite, cardClean)
		case "Unique":
			if setName == "Arthurian Legends" && slices.Contains(s.config.ALSirs, card.Name) {
				cardsElite = append(cardsElite, cardClean)
				continue
			}
			cardsUnique = append(cardsUnique, cardClean)
		default:
			return errors.New("Unknown rarity was found!")
		}
	}

	pack := getRandomCardsFromCollection(s, cardsOrdinary, cardsInPack["Ordinary"])
	pack = append(pack, getRandomCardsFromCollection(s, cardsExceptional, cardsInPack["Exceptional"])...)
	pack = append(pack, getRandomCardsFromCollection(s, cardsElite, cardsInPack["Elite"])...)
	pack = append(pack, getRandomCardsFromCollection(s, cardsUnique, cardsInPack["Unique"])...)

	fmt.Println("Random pack from all sets:")
	fmt.Println("")

	for _, card := range pack {
		fmt.Printf("%-25v | %-10v | %-15v | %-10v\n", card.Name, card.Type, card.Rarity, card.Sets)
	}

	return nil
}
*/