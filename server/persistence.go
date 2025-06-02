package server

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

type LogEntry struct {
	Op     string `json:"op"`
	ListID string `json:"list_id"`
	Value  int    `json:"value"`
}

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
	entry := LogEntry{Op: op, ListID: listID, Value: value}
	f, _ := os.OpenFile("log.jsonl", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer f.Close()
	data, _ := json.Marshal(entry)
	f.Write(append(data, '\n'))
}

func (p *PersistenceManager) StartSnapshotRoutine() {
	for {
		time.Sleep(10 * time.Second)
		p.TakeSnapshot()
	}
}

func (p *PersistenceManager) TakeSnapshot() {
	p.service.globalMu.Lock()
	defer p.service.globalMu.Unlock()

	f, _ := os.Create("snapshot.json")
	defer f.Close()
	json.NewEncoder(f).Encode(p.service.lists)
}

func (p *PersistenceManager) LoadFromSnapshotAndLog() {
	// Snapshot
	snap, err := os.ReadFile("snapshot.json")
	if err == nil {
		json.Unmarshal(snap, &p.service.lists)
	}

	// Log
	logData, err := os.ReadFile("log.jsonl")
	if err != nil {
		return
	}
	lines := splitLines(string(logData))
	for _, line := range lines {
		var entry LogEntry
		if err := json.Unmarshal([]byte(line), &entry); err == nil {
			mu := p.service.getListMutex(entry.ListID)
			mu.Lock()
			if entry.Op == "append" {
				p.service.lists[entry.ListID] = append(p.service.lists[entry.ListID], entry.Value)
			} else if entry.Op == "remove" {
				list := p.service.lists[entry.ListID]
				if len(list) > 0 {
					p.service.lists[entry.ListID] = list[:len(list)-1]
				}
			}
			mu.Unlock()
		}
	}
}

func splitLines(data string) []string {
	var lines []string
	start := 0
	for i, ch := range data {
		if ch == '\n' {
			lines = append(lines, data[start:i])
			start = i + 1
		}
	}
	if start < len(data) {
		lines = append(lines, data[start:])
	}
	return lines
}
