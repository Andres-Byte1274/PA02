package main

import (
	"fmt"
	"net"
	"net/rpc"
	"strconv"
	"sync"
	"time"
)

// Struct that is used for Tasks
type Tasks struct {
	id        int       //Id of the task
	start     int       //Used to keep track of the start of the file
	end       int       //Used to keep track of the end of the file
	status    int       //Used to get the status of the operation
	startTime time.Time //Used to track the time
}

// Struct that works as the coordinator
type Coordinator struct {
	mu sync.Mutex //used for mutual exclusion

	inputFile string //input file
	nReduce   int    //used for the reduce task
	nMap      int    // Used for the map tasks

	mapTasks    []Tasks //Used to map the tasks
	reduceTasks []Tasks //Used to reduce the tasks

	phase int // 0 = map, 1 = reduce, 2 = done

	done bool //Used to check if the task is done
}

// Used to make the coordinator
func makeCoordinator(nMap int, nReduce int, file string) *Coordinator {
	c := Coordinator{}

	c.inputFile = file
	c.nMap = nMap
	c.nReduce = nReduce
	c.phase = 0

	// Set up map tasks
	c.mapTasks = make([]Tasks, nMap)
	for i := 0; i < nMap; i++ {
		c.mapTasks[i] = Tasks{
			id:     i,
			status: 0,
		}
	}

	// Set up reduce tasks
	c.reduceTasks = make([]Tasks, nReduce)
	for i := 0; i < nReduce; i++ {
		c.reduceTasks[i] = Tasks{
			id:     i,
			status: 0,
		}
	}

	c.server()
	return &c
}

// Used to listen for RPC request
func (c *Coordinator) GetTask(req *TaskRequest, reply *TaskReply) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.phase == 2 {
		reply.TaskType = "done"
		return nil
	}

	if c.phase == 0 {
		allDone := true

		for i := 0; i < len(c.mapTasks); i++ {
			task := &c.mapTasks[i]

			if task.status != 2 {
				allDone = false
			}

			if task.status == 0 || (task.status == 1 && time.Since(task.startTime) > 10*time.Second) {
				task.status = 1
				task.startTime = time.Now()

				// Send the split file name for this map task
				reply.TaskType = "map"
				reply.TaskID = task.id
				reply.FileName = "split-" + strconv.Itoa(task.id)
				reply.NReduce = c.nReduce
				reply.NMap = c.nMap
				return nil
			}
		}

		if allDone {
			c.phase = 1
		} else {
			reply.TaskType = "wait"
			return nil
		}
	}

	if c.phase == 1 {
		allDone := true

		for i := 0; i < len(c.reduceTasks); i++ {
			task := &c.reduceTasks[i]

			if task.status != 2 {
				allDone = false
			}

			if task.status == 0 || (task.status == 1 && time.Since(task.startTime) > 10*time.Second) {
				task.status = 1
				task.startTime = time.Now()

				reply.TaskType = "reduce"
				reply.TaskID = task.id
				reply.NReduce = c.nReduce
				reply.NMap = c.nMap
				return nil
			}
		}

		if allDone {
			c.phase = 2
			c.done = true
			reply.TaskType = "done"
			return nil
		}

		reply.TaskType = "wait"
		return nil
	}

	reply.TaskType = "wait"
	return nil
}

func (c *Coordinator) ReportDone(req *DoneRequest, reply *DoneReply) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if req.TaskType == "map" {
		c.mapTasks[req.TaskID].status = 2
	} else if req.TaskType == "reduce" {
		c.reduceTasks[req.TaskID].status = 2
	}

	return nil
}

func (c *Coordinator) Done() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.done
}

func (c *Coordinator) server() {
	rpc.Register(c)

	l, err := net.Listen("tcp", ":1234")
	if err != nil {
		fmt.Println("listen error:", err)
		return
	}

	go rpc.Accept(l)
}
