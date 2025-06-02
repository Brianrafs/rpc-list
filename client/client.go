package main

import (
	"fmt"
	"log"
	"net/rpc"
	"github.com/brianrafs/rpc-list/server"
)

func main() {
	client, err := rpc.Dial("tcp", "localhost:1234")
	if err != nil {
		log.Fatal("Erro ao conectar ao servidor:", err)
	}

	var reply string
	err = client.Call("RemoteListService.Append", server.AppendArgs{ListID: "lista1", Value: 42}, &reply)
	fmt.Println("Append:", reply)

	var size int
	client.Call("RemoteListService.Size", server.SizeArgs{ListID: "lista1"}, &size)
	fmt.Println("Size:", size)

	var value int
	client.Call("RemoteListService.Get", server.GetArgs{ListID: "lista1", Index: 0}, &value)
	fmt.Println("Get index 0:", value)

	client.Call("RemoteListService.Remove", server.RemoveArgs{ListID: "lista1"}, &value)
	fmt.Println("Remove:", value)
}
