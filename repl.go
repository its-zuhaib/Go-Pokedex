package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
)

func cleanInput(text string) []string {

	lowered := strings.ToLower(text)
	return strings.Fields(lowered)
}

type cliCommand struct {
	name        string
	description string
	callback    func(*config, ...string) error
}

func getCommands() map[string]cliCommand {
	return map[string]cliCommand{
		"exit": {
			name:        "exit",
			description: " Exit the Pokedex.",
			callback:    commandExit,
		},
		"help": {
			name:        "help",
			description: "Displays a help message.",
			callback:    commandHelp,
		},
		"map": {
			name:        "map",
			description: "Displays the names of locations.",
			callback:    commandMap,
		},
		"mapb": {
			name:        "mapb",
			description: "Displays the names of previous locations.",
			callback:    commandBackMap,
		},
		"explore": {
			name:        "explore",
			description: "Displays the names of pokemons available in a particular location area.",
			callback:    commandExplore,
		},
		"catch": {
			name:        "catch",
			description: "Attempts to catch the mentioned pokemon. If successful, adds the pokemon to the pokedex.",
			callback:    commandCatch,
		},
		"inspect": {
			name:        "inspect",
			description: "List the details of the mentioned pokemon, provided its been caught by the user. ",
			callback:    commandInspect,
		},
		"pokedex": {
			name:        "pokedex",
			description: "List the details of all the pokemons caught by the user.",
			callback:    commandPokedex,
		},
	}
}

func commandExit(cfg *config, extra ...string) error {
	fmt.Printf("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp(cfg *config, extra ...string) error {
	fmt.Println("Welcome to the Pokedex!")
	fmt.Println("Usage:")
	commands := getCommands()
	for _, value := range commands {
		fmt.Printf("%s: %s\n", value.name, value.description)
	}
	return nil
}

func commandMap(cfg *config, extra ...string) error {
	url := "https://pokeapi.co/api/v2/location-area/"
	if cfg.nextURL != nil {
		url = *cfg.nextURL
	}
	var locationMap mapResponse

	data, ok := cfg.cache.Get(url)
	if ok {

		err := json.Unmarshal(data, &locationMap)
		if err != nil {
			return err
		}
	} else {
		res, err := http.Get(url)
		if err != nil {
			return err
		}
		defer res.Body.Close()

		data, err = io.ReadAll(res.Body)
		if err != nil {
			return err
		}
		cfg.cache.Add(url, data)
		err = json.Unmarshal(data, &locationMap)
		if err != nil {
			return err
		}
	}

	cfg.nextURL = locationMap.Next
	cfg.prevURL = locationMap.Previous
	for _, loc := range locationMap.Results {
		fmt.Println(loc.Name)
	}

	return nil

}

func commandBackMap(cfg *config, extra ...string) error {
	if cfg.prevURL == nil {
		fmt.Println("you're on the first page")
		return nil
	}
	url := *cfg.prevURL
	data, ok := cfg.cache.Get(url)
	var locationMap mapResponse
	if ok {
		err := json.Unmarshal(data, &locationMap)
		if err != nil {
			return err
		}
	} else {
		res, err := http.Get(url)
		if err != nil {
			return err
		}

		data, err := io.ReadAll(res.Body)
		defer res.Body.Close()
		if err != nil {
			return err
		}
		cfg.cache.Add(url, data)
		err = json.Unmarshal(data, &locationMap)
		if err != nil {
			return err
		}
	}

	cfg.nextURL = locationMap.Next
	cfg.prevURL = locationMap.Previous
	for _, loc := range locationMap.Results {
		fmt.Println(loc.Name)
	}
	return nil
}

func commandExplore(cfg *config, args ...string) error {
	if len(args) != 1 {
		fmt.Print("Kindly provide a valid location name\n")
		return fmt.Errorf("Kindly provide a valid location name")
	}

	locationName := args[0]
	u, _ := url.Parse("https://pokeapi.co/api/v2/location-area")
	u.Path = path.Join(u.Path, locationName, "/")
	fullURL := u.String()
	data, ok := cfg.cache.Get(fullURL)

	var response locationExploreResponse
	if !ok {
		res, err := http.Get(fullURL)
		if err != nil {
			return err
		}
		defer res.Body.Close()
		data, err = io.ReadAll(res.Body)
		if err != nil {
			return err
		}
		cfg.cache.Add(fullURL, data)
		err = json.Unmarshal(data, &response)
	} else {
		err := json.Unmarshal(data, &response)
		if err != nil {
			return err
		}
	}

	fmt.Printf("Exploring %s...\n", locationName)
	fmt.Println("Found Pokemon:")
	for _, encounter := range response.PokemonEncounters {
		fmt.Printf(" - %s\n", encounter.Pokemon.Name)
	}
	return nil

}

func commandCatch(cfg *config, args ...string) error {
	if len(args) != 1 {
		fmt.Printf("Enter exactly 1 valid pokemon to catch.\n")
		return fmt.Errorf("Enter only 1 pokemon to catch.")
	}
	name := args[0]

	u, _ := url.Parse("https://pokeapi.co/api/v2/pokemon/")
	u.Path = path.Join(u.Path, name, "/")
	finalURL := u.String()
	data, ok := cfg.cache.Get(finalURL)
	var pokemon Pokemon
	if !ok {
		res, err := http.Get(finalURL)
		if err != nil {
			return err
		}
		defer res.Body.Close()
		data, err = io.ReadAll(res.Body)
		cfg.cache.Add(finalURL, data)
		err = json.Unmarshal(data, &pokemon)
		if err != nil {
			fmt.Printf("Invalid pokemon!\n")
			return err
		}
	} else {
		err := json.Unmarshal(data, &pokemon)
		if err != nil {
			return err
		}
	}

	res := rand.Intn(pokemon.BaseExperience)
	fmt.Printf("Throwing a Pokeball at %s...", name)
	if res > (pokemon.BaseExperience / 2) {
		fmt.Printf("%s was caught!\nYou may now inspect it with the inspect command.\n", name)

		cfg.pokedex[name] = pokemon
	} else {
		fmt.Printf("%s escaped!\n", name)
	}
	return nil
}

func commandInspect(cfg *config, args ...string) error {
	if len(args) != 1 {
		return fmt.Errorf("you must provide a pokemon name")
	}

	name := args[0]
	pokemon, ok := cfg.pokedex[name]

	if !ok {
		fmt.Println("you have not caught that pokemon")
		return nil
	}

	fmt.Printf("Name: %s\n", pokemon.Name)
	fmt.Printf("Height: %d\n", pokemon.Height)
	fmt.Printf("Weight: %d\n", pokemon.Weight)

	fmt.Println("Stats:")
	for _, s := range pokemon.Stats {
		fmt.Printf("  -%s: %d\n", s.Stat.Name, s.BaseStat)
	}

	fmt.Println("Types:")
	for _, t := range pokemon.Types {
		fmt.Printf("  - %s\n", t.Type.Name)
	}

	return nil
}

func commandPokedex(cfg *config, args ...string) error {
	fmt.Printf("Your Pokedex:\n")
	for _, pokemon := range cfg.pokedex {
		fmt.Printf(" - %s\n", pokemon.Name)
	}
	return nil
}
