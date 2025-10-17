package main

import (
	"github.com/maxBRT/go-keeper/auth"
	"github.com/maxBRT/go-keeper/gateway"
)

func main() {
	gw := gateway.New("6969")

	authProvider, err := auth.NewAuthProvider()
	if err != nil {
		panic(err)
	}

	gw.Use(authProvider.JWTMiddleware)

	gw.Route("/*", "httpbin.org")

	if err := gw.Run(); err != nil {
		panic(err)
	}
}
