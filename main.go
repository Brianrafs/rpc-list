package main

import (
	"fmt"
	"log"
	"net"
	"net/rpc"

	"github.com/brianrafs/rpc-list/server"
)

func main() {
	rpc.Register(server.NewRemoteListService())
	listener, err := net.Listen("tcp", ":1234")
	if err != nil {
		log.Fatalf("Erro ao escutar: %v", err)
	}
	fmt.Println("Servidor RPC escutando na porta 1234...")
	rpc.Accept(listener)
}