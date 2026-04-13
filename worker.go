package main

import (
	"fmt"
	"net/rpc"
	"os"
	"strconv"
	"strings"
	"time"
)

// hash function for a string
func ihash(s string) int {
	h := 0
	for i := 0; i < len(s); i++ {
		h = h*31 + int(s[i])
	}
	if h < 0 {
		h = -h
	}
	return h
}

func startWorker() {
	for {
		// connect to coordinator w/ RPC
		client, err := rpc.Dial("tcp", "localhost:1234")
		if err != nil {
			fmt.Println("Error connecting to coordinator:", err)
			return
		}

		// call GetTask
		req := TaskRequest{}
		reply := TaskReply{}
		err = client.Call("Coordinator.GetTask", &req, &reply)
		if err != nil {
			fmt.Println("Error calling GetTask:", err)
			client.Close()
			return
		}
		client.Close()

		// handle task type
		if reply.TaskType == "map" {
			doMapTask(reply)
		} else if reply.TaskType == "reduce" {
			doReduceTask(reply)
		} else if reply.TaskType == "wait" {
			// Wait a second and try again
			time.Sleep(time.Second)
		} else if reply.TaskType == "done" {
			return
		}
	}
}

func doMapTask(reply TaskReply) {
	// read the split file
	data, err := os.ReadFile(reply.FileName)
	if err != nil {
		fmt.Println("Error reading split file:", err)
		return
	}

	// split into words
	content := string(data)
	words := strings.Fields(content)

	// create buckets for each reduce task
	buckets := make([][]string, reply.NReduce)
	for i := 0; i < reply.NReduce; i++ {
		buckets[i] = []string{}
	}

	// strip punctuation from each word
	for i := 0; i < len(words); i++ {
		clean := ""
		for j := 0; j < len(words[i]); j++ {
			c := words[i][j]
			if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') {
				clean = clean + string(c)
			}
		}
		words[i] = clean
	}

	// put each word into its bucket
	for i := 0; i < len(words); i++ {
		if words[i] == "" {
			continue
		}
		bucket := ihash(words[i]) % reply.NReduce
		buckets[bucket] = append(buckets[bucket], words[i])
	}

	// write each bucket to an intermed file
	for j := 0; j < reply.NReduce; j++ {
		// build file content
		fileContent := ""
		for i := 0; i < len(buckets[j]); i++ {
			fileContent = fileContent + buckets[j][i] + "\n"
		}

		// write intermed file
		fname := "intermediate-" + strconv.Itoa(reply.TaskID) + "-" + strconv.Itoa(j)
		err = os.WriteFile(fname, []byte(fileContent), 0644)
		if err != nil {
			fmt.Println("Error writing intermediate file:", err)
			return
		}
	}

	// tell coordinator map task is done
	client, err := rpc.Dial("tcp", "localhost:1234")
	if err != nil {
		fmt.Println("Error connecting to coordinator:", err)
		return
	}

	doneReq := DoneRequest{}
	doneReq.TaskType = "map"
	doneReq.TaskID = reply.TaskID
	doneReply := DoneReply{}
	err = client.Call("Coordinator.ReportDone", &doneReq, &doneReply)
	if err != nil {
		fmt.Println("Error calling ReportDone:", err)
		client.Close()
		return
	}
	client.Close()
}

func doReduceTask(reply TaskReply) {
	// count words from all intermed files for reduce bucket
	counts := make(map[string]int)

	// read all intermediate files for this reduce task
	for i := 0; i < reply.NMap; i++ {
		fname := "intermediate-" + strconv.Itoa(i) + "-" + strconv.Itoa(reply.TaskID)

		// read  file
		data, err := os.ReadFile(fname)
		if err != nil {
			fmt.Println("Error reading intermediate file:", err)
			return
		}

		// split by lines
		content := string(data)
		lines := strings.Split(content, "\n")

		// count each word
		for j := 0; j < len(lines); j++ {
			word := lines[j]
			if word == "" {
				continue
			}
			counts[word] = counts[word] + 1
		}
	}

	// build output string
	output := ""
	for word, count := range counts {
		output = output + fmt.Sprintf("%s %d\n", word, count)
	}

	// write output file
	fname := "mr-out-" + strconv.Itoa(reply.TaskID)
	err := os.WriteFile(fname, []byte(output), 0644)
	if err != nil {
		fmt.Println("Error writing output file:", err)
		return
	}

	// tell coordinator reduce task is done
	client, err := rpc.Dial("tcp", "localhost:1234")
	if err != nil {
		fmt.Println("Error connecting to coordinator:", err)
		return
	}

	doneReq := DoneRequest{}
	doneReq.TaskType = "reduce"
	doneReq.TaskID = reply.TaskID
	doneReply := DoneReply{}
	err = client.Call("Coordinator.ReportDone", &doneReq, &doneReply)
	if err != nil {
		fmt.Println("Error calling ReportDone:", err)
		client.Close()
		return
	}
	client.Close()
}
