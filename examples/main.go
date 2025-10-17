package main

import "github.com/maxBRT/go-keeper/gateway"

func main() {
	gw := gateway.New("6969")

	gw.Route("/status/*", "httpbin.org")

	if err := gw.Run(); err != nil {
		panic(err)
	}

}
