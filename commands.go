package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/uller91/sorceryPackGen/internal/apiInter"
	"github.com/uller91/sorceryPackGen/internal/database"
	"slices"
	"time"
)

type state struct {
	config   *config
	database *database.Queries
	commands *commands
}

type command struct {
	name      string
	arguments []string
}

type commands struct {
	handlers     map[string]func(*state, command) error
	descriptions map[string]string
}

func (c *commands) run(s *state, cmd command) error {
	hndl, exists := c.handlers[cmd.name]
	if exists {
		err := hndl(s, cmd)
		if err != nil {
			return err
		}
	} else {
		return errors.New("No command with this name is registered")
	}

	return nil
}

func (c *commands) register(name string, f func(*state, command) error, d string) {
	c.handlers[name] = f
	c.descriptions[name] = d
}

func addToCollection(origin *[]string, collection *[]string, item string) {
	if !slices.Contains(*origin, item) && !slices.Contains(*collection, item) {
		*collection = append(*collection, item)
	}
}

// Version

const (
	descriptionVersion = "Shows the current version of the app"
)

func handlerVersion(s *state, cmd command) error {
	if len(cmd.arguments) != 0 {
		return errors.New("0 arguments are expected")
	}

	fmt.Println("1.0.0")
	return nil
}

// Help
const (
	descriptionHelp = "Shows the list of commands (\"help\") or their description (\"help command_name\")"
)

func handlerHelp(s *state, cmd command) error {
	if len(cmd.arguments) == 0 {
		fmt.Println("List of available commands:")
		fmt.Println("")
		commandNames := []string{}
		for command, _ := range s.commands.handlers {
			commandNames = append(commandNames, command)
		}
		slices.Sort(commandNames)
		for _, name := range commandNames {
			fmt.Println(name)
		}
		fmt.Println("")
		fmt.Println("To show the command's description use \"help command_name\"")
	} else if len(cmd.arguments) == 1 {
		if discription, ok := s.commands.descriptions[cmd.arguments[0]]; ok {
			fmt.Printf("Command \"%s\"\n", cmd.arguments[0])
			fmt.Println("")
			fmt.Println(discription)
		} else {
			return errors.New("No command with this name is registered")
		}
	} else {
		return errors.New("Too many arguments")
	}

	return nil
}

// Sets
const (
	descriptionSets = "Shows the list of sets available in the DB"
)

func handlerSets(s *state, cmd command) error {
	if len(cmd.arguments) != 0 {
		return errors.New("0 arguments are expected")
	}

	err := s.updateConfig()
	if err != nil {
		return err
	}

	if len(s.config.Sets) == 0 {
		fmt.Println("The DB is empty! Use \"update\" command to fill the DB with cards")
	} else {
		fmt.Println("List of sets in the DB:")
		fmt.Println("")
		for _, set := range s.config.Sets {
			fmt.Println(set)
		}
	}

	return nil
}

// Update
const (
	descriptionUpdate = "Updates an internal card DB by sending the API requiest to \"api.sorcerytcg.com\""
)

// current db size = 649
func handlerUpdate(s *state, cmd command) error {
	if len(cmd.arguments) != 0 {
		return errors.New("0 arguments are expected")
	}

	err := s.updateConfig()
	if err != nil {
		return err
	}

	var newSets []string

	fmt.Println("Initializing card DB update...")
	fmt.Println("")
	fmt.Printf("Sending the API requiest to %v...\n", s.config.BaseUrl)
	fmt.Println("")

	cards, err := apiInter.RequestCard(s.config.BaseUrl)
	if err != nil {
		return err
	}

	dbSize := len(cards)
	fmt.Printf("Cards found: %v\n", dbSize)
	fmt.Println("")

	fmt.Println("Updating card DB...")

	cardsAdded := 0
	cardsUpdated := 0

	for _, card := range cards {
		//add card
		paramCard := database.CreateCardParams{ID: uuid.New(), CreatedAt: time.Now(), UpdatedAt: time.Now(), Name: card.Name, Rarity: card.Guardian.Rarity, Type: card.Guardian.Type}
		cardCreated, err := s.database.CreateCard(context.Background(), paramCard)
		if err != nil {
			if pqError, ok := err.(*pq.Error); ok && pqError.Code == "23505" {
				cardsUpdated += 1
			} else {
				return err
			}
		} else {
			cardsAdded += 1
		}

		//add set+card
		for _, set := range card.Sets {
			paramSets := database.CreateSetAndCardParams{ID: uuid.New(), CreatedAt: time.Now(), UpdatedAt: time.Now(), Name: set.Name, CardID: cardCreated.ID}
			_, err := s.database.CreateSetAndCard(context.Background(), paramSets)

			if err != nil {
				pqError, ok := err.(*pq.Error)
				if ok && pqError.Code == "23503" {
					continue
				} else {
					return err
				}
			}
			addToCollection(&s.config.Sets, &newSets, set.Name)
		}
	}

	//add new sets
	if newSets != nil {
		fmt.Println("")
		fmt.Println("New sets added to the DB:")

		for _, set := range newSets {
			paramSet := database.CreateSetParams{ID: uuid.New(), CreatedAt: time.Now(), UpdatedAt: time.Now(), Name: set}
			_, err := s.database.CreateSet(context.Background(), paramSet)
			if err != nil {
				return err
			}
			fmt.Println(set)
		}
	}

	fmt.Println("")
	fmt.Printf("Cards added in the DB: %v\n", cardsAdded)
	fmt.Printf("Cards already in the DB: %v\n", cardsUpdated)
	fmt.Println("")

	fmt.Println("SorceryDB update finished successfully")

	return nil
}

// Reset
const (
	descriptionReset = "Deletes all the entries from the DB"
)

func handlerReset(s *state, cmd command) error {
	if len(cmd.arguments) != 0 {
		return errors.New("0 arguments are expected")
	}

	err := s.database.SetlistReset(context.Background())
	if err != nil {
		return err
	}

	err = s.database.SetsReset(context.Background())
	if err != nil {
		return err
	}

	err = s.database.CardsReset(context.Background())
	if err != nil {
		return err
	}

	fmt.Println("Card DB successfully reset")

	return nil
}

// Open
const (
	descriptionGenerate = `Generates (prints) a number of random card packs from Sorcery TCG
By default generates one standard pack (15 cards: 11 Ordinary, 3 Exceptional, and 1 Elite (or Unique instead with 20% chance))

Available tags:
-n number
	Changes the number of packs to generate. Only accepts positive integers.
	If number > 1 the program will wait before displaying each pack (click Enter to open a pack).
	When multiple packs are generated it is considered that you have a "full collection",
	i.e. 4x of each Ordinary, 3x of each Exceptional, 2x of each Elite and 1x of each Unique card.
	The program can't generate more packs that there are cards available in the collection.
	In case of multiple packs generated, all of them will have the same pack content (see -p tag).

-s "set_name"
	Changes the set from which packs are generated. Accepts full name of the set (you can use "sets" command to see the full list)
	or their abbreviation as well as some additional variants:
		a
			"Alpha"
		b	
			"Beta"
		al	
			"Arthurian Legends"
		d	
			"Dragonlord" (can't be used at the moment since it consists of only unique cards)
		g
			"Gothic"
		p
			"Promotional" (can't be used at the moment)
		"All" 
			generates the pack from all the cards available in the DB. Expect pure lack of synergy. But it will be fun!
		"Random" and anything else
			generates a pack from the random sets (similar to not using -s tag at all)

-p "pack_type"
	Changes the content of the pack (number of cards of different rarities).
	Available configurations:
		Standard
			Standard Sorcery TCG pack (11 Ordinary, 3 Exceptional, and 1 Elite (or Unique instead with 20% chance))
		Pauper
			Pack which only includes Ordinary and Exceptional (11 Ordinary, 4 Expectional)
		Pyramid
			Pack with the more even (to rarity) cards distribution (8 Ordinary, 4 Exceptional, 2 Elite, 1 Unique)
		Ordinary, Exceptional, Elite, Unique
			15 cards pack with only given rarity
		Custom
			Allows you to set the sumber of Ordinary, Exceptional, Elite, and Unique cards in thep pack
			The program will ask you to write the number of cards of each corresponding rarity. Only works with integers.
		Random and anything else
			15 cards pack with the random number of cards of different rarities

-f	
	Adds the possibility of the "foil" card in the pack.
	In terms of this program, this tag adds 25% chance of one ordinary card to be "foil" of any rarity.

-a
	Adds the avatar cards as a first card in the pack.
	The rarity of the avatar is random since the API doesn't hold this information anymore

The following "Promotional" cards are excluded from generated packs:
	Alpha: Relentless Crowd, Winter River, Erik's Curiosa
	Draft pack: Spellslinger, Spire, Stream, Valley, Wasteland
	
One additional thing to consider. AL set has a set of Unique Sir/Dame cards which are actually Elite in terms of rarity.
Here they are considered to have Elite rarity. 
	List of Elite Sir/Dame cards:
		Dame Britomart, Sir Agravaine, Sir Balin, Sir Bedivere, Sir Bors the Younger, Sir Gaheris, Sir Gawain, Sir Ironside,
		Sir Kay, Sir Lamorak, Sir Morien, Sir Pelleas, Sir Perceval, Sir Priamus, Sir Tom Thumb, Sir Tristan`
)

func handlerGenerate(s *state, cmd command) error {
	err := s.updateConfig()
	if err != nil {
		return err
	}

	if s.config.Sets == nil {
		return errors.New("The DB is empty! Use \"update\" command to fill the DB with cards")
	}

	//setting the set, tag -s
	set, err := setSet(s, cmd)
	if err != nil {
		return err
	}

	//setting number of cards in the pack, tag -p
	cardsInPack, err := setPack(cmd)
	if err != nil {
		return err
	}

	//"foils" in the pack, tag -f
	cardsInPack, _ = setFoil(cardsInPack, cmd)

	//avatar in the pack, tag -a
	cardsInPack, _ = setAvatar(cardsInPack, cmd)

	//setting the number of packs
	numberOfPacks, err := setNumber(cmd)
	if err != nil {
		return err
	}

	if set == "All" {
		return generateMultiplePacksAll(s, cardsInPack, numberOfPacks)
		//return generateOnePackAll(s, cardsInPack)
	} else {
		return generateMiltiplePacks(s, set, cardsInPack, numberOfPacks)
		//return generateOnePack(s, set, cardsInPack)
	}

	return nil
}
