package main

import (
	"github.com/maxBRT/feather"
)

func main() {
	gw, err := feather.New("config.yaml")
	if err != nil {
		panic(err)
	}

	gw.Run("8080")
}
