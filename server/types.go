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