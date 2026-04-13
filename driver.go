package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {

	fmt.Println("Hello world!")

	// makes sure there are 4 arguments
	if len(os.Args) != 4 {
		fmt.Println("Usage: program M R inputfile")
		return
	}

	// get first command line argument and convert to int
	nMap, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Println("Invalid nMap:", err)
		return
	}

	// Error checking
	if nMap < 1 {
		fmt.Println("M cannot be less than 1")
		return
	}

	// get second command line arg and convert to int
	nReduce, err := strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Println("Invalid nReduce:", err)
		return
	}

	// error checking
	if nReduce < 1 {
		fmt.Println("R cannot be less than 1")
		return
	}

	filename := os.Args[3] // get file name

	fmt.Println("This the first argument ", nMap)
	fmt.Println("This the second argument ", nReduce)
	fmt.Println("This the third argument ", filename)

	// read input file
	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	// split into words so we don't cut a word in half
	words := strings.Fields(string(data))
	wordsPerSplit := len(words) / nMap

	// split file into nMap split files
	for i := 0; i < nMap; i++ {
		start := i * wordsPerSplit
		end := start + wordsPerSplit

		if i == nMap-1 {
			end = len(words)
		}

		// join the words for this split with spaces
		fileContent := strings.Join(words[start:end], " ")

		// write split file
		err = os.WriteFile("split-"+strconv.Itoa(i), []byte(fileContent), 0644)
		if err != nil {
			fmt.Println("Error writing split file:", err)
			return
		}

		fmt.Println("Created split file split-" + strconv.Itoa(i))
	}

	// start coordinator
	coordinator := makeCoordinator(nMap, nReduce, filename)
	if coordinator == nil {
		return
	}

	// start workers
	for i := 0; i < 3; i++ {
		go startWorker()
	}

	// wait for job to finish
	for !coordinator.Done() {
		time.Sleep(time.Second)
	}

	fmt.Println("Job finished")
}
