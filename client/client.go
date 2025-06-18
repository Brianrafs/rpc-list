package main

import (
	"fmt"
	"log"
	"net/rpc"
	"sync"
	"github.com/brianrafs/rpc-list/server"
)

func main() {
	var wg sync.WaitGroup

	// Número de clientes simulados
	numClients := 500

	for i := range make([]struct{}, numClients) {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			client, err := rpc.Dial("tcp", "localhost:1234")
			if err != nil {
				log.Printf("[Client %d] Erro ao conectar: %v", id, err)
				return
			}
			defer client.Close()

			listID := fmt.Sprintf("lista%d", id%2)
			var reply string

			// Create
			err = client.Call("RemoteListService.CreateList", server.CreateArgs{ListID: listID}, &reply)
			if err == nil {
				log.Printf("[Client %d] Lista criada: %s", id, reply)
			}

			// Append
			valueToAdd := id * 10
			err = client.Call("RemoteListService.Append", server.AppendArgs{ListID: listID, Value: valueToAdd}, &reply)
			if err != nil {
				log.Printf("[Client %d] Append error: %v", id, err)
				return
			}
			fmt.Printf("[Client %d] Append: %s\n", id, reply)

			// Append do segundo valor
			valueToAdd += 10
			err = client.Call("RemoteListService.Append", server.AppendArgs{ListID: listID, Value: valueToAdd}, &reply)
			if err != nil {
				log.Printf("[Client %d] Append error: %v", id, err)
				return
			}
			fmt.Printf("[Client %d] Append: %s\n", id, reply)

			// Get Size
			var size int
			err = client.Call("RemoteListService.Size", server.SizeArgs{ListID: listID}, &size)
			if err == nil {
				fmt.Printf("[Client %d] Size: %d\n", id, size)
			}

			// Get Index 0
			var value int
			err = client.Call("RemoteListService.Get", server.GetArgs{ListID: listID, Index: 0}, &value)
			if err == nil {
				fmt.Printf("[Client %d] Get index 0: %d\n", id, value)
			}

			// Remove último elemento
			err = client.Call("RemoteListService.Remove", server.RemoveArgs{ListID: listID}, &value)
			if err == nil {
				fmt.Printf("[Client %d] Remove: %d\n", id, value)
			}
		}(i)
	}


	// Espera todos os clientes finalizarem
	wg.Wait()
	fmt.Println("Todos os clientes terminaram.")
}
