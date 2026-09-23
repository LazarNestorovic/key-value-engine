package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/LazarNestorovic/key-value-engine/internal/config"
	"github.com/LazarNestorovic/key-value-engine/internal/engine"
)

func main() {

	db := engine.NewDBMap()
	cfg, err := config.LoadConfig("config.json")
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return
	}
	fmt.Printf("%+v\n", cfg)
	Run(db)
}

func Run(db engine.DB) {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Println("Enter command: ")
		scanner.Scan()
		line := scanner.Text()
		parts := strings.SplitN(line, " ", 2)
		operation := parts[0]
		var keyValue string
		if len(parts) == 2 {
			keyValue = parts[1]
		}
		switch operation {
		case "put":
			kv := strings.SplitN(keyValue, " ", 2)
			if len(kv) != 2 || kv[0] == "" {
				fmt.Println("Invalid command")
				continue
			}
			key, value := kv[0], kv[1]
			err := db.Put(key, []byte(value))
			if err != nil {
				fmt.Fprintln(os.Stderr, "error:", err)
				continue
			}
		case "get":
			if keyValue == "" {
				fmt.Println("Invalid command")
				continue
			}
			key := keyValue
			value, found, err := db.Get(key)
			if err != nil {
				fmt.Fprintln(os.Stderr, "error:", err)
				continue
			}
			if found {
				fmt.Println("value: " + string(value))
			} else {
				fmt.Println("There is no value for key: " + key)
			}
		case "delete":
			if keyValue == "" {
				fmt.Println("Invalid command")
				continue
			}
			key := keyValue
			_, existed, err := db.Get(key)
			if err != nil {
				fmt.Fprintln(os.Stderr, "error:", err)
				continue
			}
			if !existed {
				fmt.Println("There is no value for key: " + key)
				continue
			}
			if err := db.Delete(key); err != nil {
				fmt.Fprintln(os.Stderr, "error:", err)
				continue
			}
			fmt.Println("Item with key: " + key + " was deleted")
		case "exit":
			return
		default:
			fmt.Println("Invalid command")
		}
	}

}
