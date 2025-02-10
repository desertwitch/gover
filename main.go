package main

import (
	"fmt"
)

func main() {
	system, err := establishSystem()
	if err != nil {
		fmt.Printf("Error: %+v", err)
	}
}
