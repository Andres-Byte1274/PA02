# MapReduce Word Count

## How to Build
go build -o mapreduce .

## How to Run
./mapreduce M R inputfile
or skip build and run : go run . M R input file

Example:
./mapreduce 9 3 data.txt
 or without build : go run . 9 3 input.txt

This will:
1. Split data.txt into 9 split files (split-0 to split-8)
2. Start 1 coordinator and 3 workers
3. Produce output files mr-out-0, mr-out-1, mr-out-2

## Files
- driver.go - Main function, splits input, starts coordinator and workers
- coordinator.go - Coordinator that assigns tasks to workers
- worker.go - Worker that does map and reduce tasks
- rpc.go - Shared RPC structs

## Notes
- The coordinator listens on tcp port 1234
- Workers connect to localhost:1234
- If a worker takes more than 10 seconds, the task is reassigned
- Map tasks must all finish before reduce tasks start
- remove old outputs by using
    rm .\mr-out-*
    rm .\intermediate-*-*
    rm .\mapreduce.exe
    rm .\split-*
