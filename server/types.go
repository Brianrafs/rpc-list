package server

type AppendArgs struct {
	ListID string
	Value  int
}

type GetArgs struct {
	ListID string
	Index  int
}

type RemoveArgs struct {
	ListID string
}

type SizeArgs struct {
	ListID string
}

type CreateArgs struct {
	ListID string
}

type SnapshotFile struct {
    Timestamp int64                      `json:"timestamp"`
    Lists     map[string][]int           `json:"lists"`
}

type LogEntry struct {
    Op        string `json:"op"`
    ListID    string `json:"list_id"`
    Value     int    `json:"value"`
    Timestamp int64  `json:"timestamp"`
}
