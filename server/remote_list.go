package server

import (
	"errors"
	"sync"
)

type RemoteListService struct {
	lists map[string][]int
	mutex map[string]*sync.RWMutex
	globalMu sync.Mutex
	persist *PersistenceManager
}


func NewRemoteListService() *RemoteListService {
	s := &RemoteListService{
		lists: make(map[string][]int),
		mutex: make(map[string]*sync.RWMutex),
	}
	s.persist = NewPersistenceManager(s)
	s.persist.LoadFromSnapshotAndLog()
	stopCh := make(chan struct{})
	go s.persist.StartSnapshotRoutine(stopCh)
	return s
}

func (s *RemoteListService) getListMutex(listID string) *sync.RWMutex {
	s.globalMu.Lock()
	defer s.globalMu.Unlock()

	if s.mutex == nil {
		s.mutex = make(map[string]*sync.RWMutex)
	}
	if _, exists := s.mutex[listID]; !exists {
		s.mutex[listID] = &sync.RWMutex{}
	}
	return s.mutex[listID]
}


func (s *RemoteListService) CreateList(args CreateArgs, reply *string) error {
    s.globalMu.Lock()
    defer s.globalMu.Unlock()

    if _, exists := s.lists[args.ListID]; exists {
        return errors.New("lista já existe")
    }

    s.lists[args.ListID] = []int{}
    s.mutex[args.ListID] = &sync.RWMutex{}
    s.persist.AppendLog("create", args.ListID, 0)
    *reply = "Lista criada com sucesso"
    return nil
}

func (s *RemoteListService) Append(args AppendArgs, reply *string) error {
	mu := s.getListMutex(args.ListID)
	mu.Lock()
	defer mu.Unlock()

	_, ok := s.lists[args.ListID]

	if !ok {
		return errors.New("não é possível adicionar valor na lista, lista inexistente")
	}

	s.lists[args.ListID] = append(s.lists[args.ListID], args.Value)
	s.persist.AppendLog("append", args.ListID, args.Value)
	*reply = "Valor adicionado."
	return nil
}

func (s *RemoteListService) Get(args GetArgs, reply *int) error {
	mu := s.getListMutex(args.ListID)
	mu.RLock()
	defer mu.RUnlock()

	list, ok := s.lists[args.ListID]
	if !ok || args.Index < 0 || args.Index >= len(list) {
		return errors.New("índice inválido ou lista inexistente")
	}
	*reply = list[args.Index]
	return nil
}

func (s *RemoteListService) Remove(args RemoveArgs, reply *int) error {
	mu := s.getListMutex(args.ListID)
	mu.Lock()
	defer mu.Unlock()

	list, ok := s.lists[args.ListID]
	if !ok || len(list) == 0 {
		return errors.New("lista inexistente ou vazia")
	}
	val := list[len(list)-1]
	s.lists[args.ListID] = list[:len(list)-1]
	s.persist.AppendLog("remove", args.ListID, val)
	*reply = val
	return nil
}

func (s *RemoteListService) Size(args SizeArgs, reply *int) error {
	mu := s.getListMutex(args.ListID)
	mu.RLock()
	defer mu.RUnlock()

	list, ok := s.lists[args.ListID]
	if !ok {
		*reply = 0
		return nil
	}
	*reply = len(list)
	return nil
}