package main

// struct for task request
type TaskRequest struct {
}

// struct for task reply
type TaskReply struct {
	TaskType string // "map", "reduce", "wait", "done"
	TaskID   int
	FileName string
	Start    int
	End      int
	NReduce  int
	NMap     int
}

// struct for done request
type DoneRequest struct {
	TaskType string
	TaskID   int
}

// struct for done reply
type DoneReply struct {
}
