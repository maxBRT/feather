package main

import (
	"github.com/maxBRT/go-keeper"
)

func main() {
	gw, err := gokeeper.New("config.yaml")
	if err != nil {
		panic(err)
	}

	gw.Run("8080")
}
