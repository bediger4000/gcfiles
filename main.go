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

	// use a ticker so that time consumed outside of waiting
	// 60 seconds doesn't build up.
	tkr := time.Tick(60 * time.Second)

	for i := 0; true; i++ {
		go openAFile(i)
		_ = <-tkr
		if ((i + 1) % 10) == 0 {
			fmt.Printf("%s - created %d files, collecting garbage\n",
				time.Since(start), i+1)
			runtime.GC()
		}
		n := countFileDescriptors(os.Getpid())
		fmt.Printf("%s - created %d files, %d open file descriptors likely from creating files\n",
			time.Since(start), i+1, n)
	}
}

// openAFile - creates a *os.File struct representing a file in the
// current working directory. Does not call (*os.File).Close() on that
// struct, but the memory allocation for the *os.File should be
// eligible for garbage collection after this function returns.
func openAFile(n int) {
	f, err := os.OpenFile(fmt.Sprintf("%s%04d", fileNamePrefix, n), os.O_RDWR|os.O_CREATE, 0777)
	if err != nil {
		log.Printf("creating file %d: %v\n", n, err)
		return
	}
	// put something in the file
	fmt.Fprintf(f, "contents of file %d\n", n)
}

// Name the files created by func openAFile something unique, so
// that a regexp can recognize those filenames from the file desriptors
// enumerated in /proc/$PID/fd/
const fileNamePrefix = "Glump"

var fileNamePattern = regexp.MustCompile(fmt.Sprintf("^/..*/%s[0-9]+$", fileNamePrefix))

// countFileDescriptors looks at files in /proc/$PID/fd.
// It follows them (they're links) to get the fully-qualified
// path of the file represented by each file descriptor.
func countFileDescriptors(pid int) int {

	entries, err := os.ReadDir(fmt.Sprintf("/proc/%d/fd", pid))
	if err != nil {
		log.Print(err)
		return -1
	}

	count := 0
	// each entry in /proc/$PID/fd/ is something like "1", "2", "3", etc
	// and each entry is almost certainly a link. The linked file's path
	// gets examined to see if that file was created by os.CreateTemp in
	// func openAFile, and counted if so.
	for i := range entries {
		path, err := os.Readlink(fmt.Sprintf("/proc/%d/fd/%s", pid, entries[i].Name()))
		if err != nil {
			// Don't bother to output an error message, it's not worth it.
			continue
		}
		if fileNamePattern.MatchString(path) {
			count++
		}
	}
	return count
}
