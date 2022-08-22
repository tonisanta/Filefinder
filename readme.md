# Filefinder
This library allows you to find the files containing a particular string. It returns 
a channel, as the traverse of the directories is done concurrently, as soon as 
a file is found, it's inserted in the channel.

```go
func ExampleFindFiles() {
	outChan, err := FindFiles(testFS, "Hello")
	if err != nil {
		log.Fatalf("Error: %s", err)
	}

	for file := range outChan {
		fmt.Println(file)
	}

	// Keep in mind that the order is not deterministic
	// Output: greeting.txt
	// subdirectory\greeting2.txt
	// subdirectory\another\greeting3.txt
}

```

## Benchmarking

The standard library provides a function to traverse a directory (_WalkDir_), and you
can provide a custom function to run for each entry, however, this traverse is done
in a single thread.

On the other hand, this function creates a goroutine per each entry, this way 
multiple files can be checked at the same time. This way, as soon as a file is found, 
the consumer can start doing any other process with it.


| Function               | Mean execution time | Bytes allocated | Number allocations |
|------------------------|--------------------:|----------------:|-------------------:|
| BenchmarkWalkDirStdLib |     131338412 ns/op |  526225622 B/op |  1032780 allocs/op |
| BenchmarkWalkDir       |      55356895 ns/op |  410276759 B/op |  1017930 allocs/op |