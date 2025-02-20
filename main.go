package main

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"runtime"
	"time"
)

func main() {

	start := time.Now()

	tkr := time.Tick(60 * time.Second)

	for i := 0; true; i++ {
		go openAFile(i)
		_ = <-tkr
		if ((i + 1) % 10) == 0 {
			fmt.Printf("Created %d files, collecting garbage\n", i+1)
			runtime.GC()
		}
		n := countFileDescriptors(os.Getpid())
		fmt.Printf("%s - created %d files, %d open file descriptors likely from creating files\n",
			time.Since(start), i+1, n)
	}
}

const fileNamePrefix = "Glump"

func openAFile(n int) {
	f, err := os.OpenFile(fmt.Sprintf("%s%04d", fileNamePrefix, n), os.O_RDWR|os.O_CREATE, 0777)
	if err != nil {
		log.Printf("creating file %d: %v\n", n, err)
		return
	}
	// put something in the file
	fmt.Fprintf(f, "contents of file %d\n", n)
}

var fileNamePattern = regexp.MustCompile(fmt.Sprintf("^/..*/%s[0-9]+$", fileNamePrefix))

func countFileDescriptors(pid int) int {

	entries, err := os.ReadDir(fmt.Sprintf("/proc/%d/fd", pid))
	if err != nil {
		log.Print(err)
		return -1
	}

	count := 0
	for i := range entries {
		path, err := os.Readlink(fmt.Sprintf("/proc/%d/fd/%s", pid, entries[i].Name()))
		if err != nil {
			// log.Printf("%s not a link or something: %v\n", entries[i].Name(), err)
			continue
		}
		if fileNamePattern.MatchString(path) {
			count++
		}
	}
	return count
}
