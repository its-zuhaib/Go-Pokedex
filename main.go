package main

import (
	"bufio"
	"fmt"

	"os"
	"strings"
	"time"

	"github.com/its-zuhaib/go_pokedex/internal/pokecache"
)

func main() {

	myCache := pokecache.NewCache(5 * time.Minute)

	cfg := &config{
		nextURL: nil,
		prevURL: nil,
		cache:   myCache,
		pokedex: make(map[string]Pokemon),
	}

	// Start the CLI
	scanner := bufio.NewScanner(os.Stdin)
	for i := 0; ; i++ {
		fmt.Print("Pokedex > ")

		for scanner.Scan() {
			userInput := scanner.Text()
			userInput = strings.TrimSpace(userInput)

			words := strings.Fields(strings.ToLower(userInput))

			command := words[0]
			args := words[1:]
			commands := getCommands()
			value, ok := commands[command]
			if !ok {
				fmt.Printf("Unknown command\n")
				fmt.Print("Pokedex > ")
				continue
			}
			value.callback(cfg, args...)
			fmt.Print("Pokedex > ")
		}
	}
}
