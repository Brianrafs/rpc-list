package server

import (
	"encoding/json"
	"log"
	"os"
	"sync"
	"time"
)

type PersistenceManager struct {
	service *RemoteListService
	logMu   sync.Mutex
}

func NewPersistenceManager(s *RemoteListService) *PersistenceManager {
	return &PersistenceManager{service: s}
}

func (p *PersistenceManager) AppendLog(op, listID string, value int) {
	p.logMu.Lock()
	defer p.logMu.Unlock()

	entry := LogEntry{
		Op:        op,
		ListID:    listID,
		Value:     value,
		Timestamp: time.Now().UnixNano(),
	}

	f, err := os.OpenFile("log.jsonl", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("erro ao abrir o arquivo de log: %v", err)
		return
	}
	defer func() {
		if cerr := f.Close(); cerr != nil {
			log.Printf("erro ao fechar o arquivo de log: %v", cerr)
		}
	}()

	data, err := json.Marshal(entry)
	if err != nil {
		log.Printf("erro ao serializar log: %v", err)
		return
	}

	if _, err := f.Write(append(data, '\n')); err != nil {
		log.Printf("erro ao gravar no log: %v", err)
	}
}

// StartSnapshotRoutine inicia uma rotina em background para tirar snapshots periodicamente.
func (p *PersistenceManager) StartSnapshotRoutine(stopCh <-chan struct{}) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("panic na rotina de snapshot: %v", r)
			}
		}()
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				log.Println("Executando snapshot automático...")
				p.TakeSnapshot()
			case <-stopCh:
				log.Println("Rotina de snapshot encerrada.")
				return
			}
		}
	}()
}

func (p *PersistenceManager) TakeSnapshot() {
	p.service.globalMu.Lock()
	defer p.service.globalMu.Unlock()

	data := SnapshotFile{
		Timestamp: time.Now().UnixNano(),
		Lists:     make(map[string][]int, len(p.service.lists)),
	}

	// Deep copy das listas para evitar race conditions
	for k, v := range p.service.lists {
		copied := make([]int, len(v))
		copy(copied, v)
		data.Lists[k] = copied
	}

	f, err := os.Create("snapshot.json")
	if err != nil {
		log.Printf("erro ao criar snapshot: %v", err)
		return
	}
	defer func() {
		if cerr := f.Close(); cerr != nil {
			log.Printf("erro ao fechar snapshot: %v", cerr)
		}
	}()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(data); err != nil {
		log.Printf("erro ao serializar snapshot: %v", err)
	}
}

func (p *PersistenceManager) LoadFromSnapshotAndLog() {
	var snapshotTime int64

	// --- Lê o snapshot ---
	snap, err := os.ReadFile("snapshot.json")
	if err == nil {
		var snapData SnapshotFile
		if err := json.Unmarshal(snap, &snapData); err == nil {
			// Deep copy para evitar race conditions
			newLists := make(map[string][]int, len(snapData.Lists))
			for k, v := range snapData.Lists {
				copied := make([]int, len(v))
				copy(copied, v)
				newLists[k] = copied
			}
			p.service.lists = newLists
			snapshotTime = snapData.Timestamp
		} else {
			log.Printf("erro ao desserializar snapshot: %v", err)
		}
	} else if !os.IsNotExist(err) {
		log.Printf("erro ao ler snapshot: %v", err)
	}

	// --- Lê e aplica apenas logs mais recentes que o snapshot ---
	logData, err := os.ReadFile("log.jsonl")
	if err != nil {
		if !os.IsNotExist(err) {
			log.Printf("erro ao ler log: %v", err)
		}
		return
	}

	lines := splitLines(string(logData))
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		var entry LogEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			log.Printf("erro ao desserializar linha do log: %v", err)
			continue
		}
		// ⛔ Pula logs antigos
		if entry.Timestamp <= snapshotTime {
			continue
		}

		switch entry.Op {
		case "create":
			if _, exists := p.service.lists[entry.ListID]; !exists {
				p.service.lists[entry.ListID] = []int{}
				p.service.mutex[entry.ListID] = &sync.RWMutex{}
			}
		case "append":
			mu := p.service.getListMutex(entry.ListID)
			mu.Lock()
			p.service.lists[entry.ListID] = append(p.service.lists[entry.ListID], entry.Value)
			mu.Unlock()
		case "remove":
			mu := p.service.getListMutex(entry.ListID)
			mu.Lock()
			list := p.service.lists[entry.ListID]
			if len(list) > 0 {
				p.service.lists[entry.ListID] = list[:len(list)-1]
			}
			mu.Unlock()
		default:
			log.Printf("operação desconhecida no log: %s", entry.Op)
		}
	}
}

func splitLines(data string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(data); i++ {
		if data[i] == '\n' {
			line := data[start:i]
			
			if len(line) > 0 && line[len(line)-1] == '\r' {
				line = line[:len(line)-1]
			}
			lines = append(lines, line)
			start = i + 1
		}
	}
	
	if start < len(data) {
		line := data[start:]
		if len(line) > 0 && line[len(line)-1] == '\r' {
			line = line[:len(line)-1]
		}
		lines = append(lines, line)
	}
	return lines
}
