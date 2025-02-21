# Go Runtime Garbage Collects Open Files

It turns out that the Go standard library hides Linux file descriptors
(small integer values representing open files) from programmers
via the standard package `os`, as type `*os.File`.
Type  `*os.File` defines a "finalizer" that closes the open file descriptor
when the memory of the `*os.File` gets garbage collected.

I wrote a program to prove and test this.

## Build and run

After cloning the repository, change to that directory, might be `$GOPATH/src/gcfiles`.

```
$ go build $PWD
$ ./gcfiles
1m0.059950773s - created 1 files, 1 open file descriptors likely from creating files
2m0.060334339s - created 2 files, 2 open file descriptors likely from creating files
3m0.060723467s - created 3 files, 3 open file descriptors likely from creating files
4m0.060062449s - created 4 files, 4 open file descriptors likely from creating files
5m0.028245944s - created 5 files, 5 open file descriptors likely from creating files
6m0.060777374s - created 6 files, 6 open file descriptors likely from creating files
7m0.060233541s - created 7 files, 7 open file descriptors likely from creating files
8m0.060617395s - created 8 files, 8 open file descriptors likely from creating files
9m0.013064547s - created 9 files, 9 open file descriptors likely from creating files
10m0.060166522s - created 10 files, collecting garbage
10m0.063337033s - created 10 files, 0 open file descriptors likely from creating files
11m0.001096269s - created 11 files, 1 open file descriptors likely from creating files
12m0.018551404s - created 12 files, 2 open file descriptors likely from creating files
13m0.002885157s - created 13 files, 0 open file descriptors likely from creating files
...
```

### What does it all mean?

The program loops indefinitely.
Once a minute, it runs a goroutine (a.k.a. thread) that 
creates an open file via the `os.CreateTemp` function.
`os.CreateTemp` returns a `*os.File` pointer.
The goroutine writes into the file via that pointer,
and exits.
The goroutine does not close the file via the `(*os.File).Close()`
method.
That leaves the file descriptor represented by the `*os.File` open.

You can see a Linux process' open file descriptors in a pseudo-directory,
`/proc/$PID/fd/`.
Try it (you are using Linux, right?): `ls -l /proc/$$/fd`

The program reads file names from the directory `/proc/$PID/fd`
(`$PID` represents the program's process ID when it's running).
File names are links, and look like "1", "2", "3" ...
The linked-to files in the case of this program are files
created by `os.CreateTemp`.
As the program runs, it racks up open file descriptors.
Remember`that the program does not call (*os.File).Close()`.
Every tenth file (conveniently, every 10 minutes),
the program calls `runtime.GC()`, triggering a garbage collection.
Since 10 open file descriptors exist at this point,
the "finalizer" set up in the `os` package code gets executed,
and the encapsulated file descriptors get closed.

I believed that the Go runtime did a garbage collection after
60 seconds of execution, or some other threshold amount of allocations
got hit, whichever came first.
That's why this program creates a new file every 60 seconds.
Apparently, I was wrong.
