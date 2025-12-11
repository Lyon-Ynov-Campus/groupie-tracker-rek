package main

import (
	"net"
)

type GameServer struct {
	Letter    byte
	Categorie string
}

func ptitbacHandler(letter byte, categorie string) *GameServer {