stringio
--------

A stringio library for Go (like the python StringIO). Forked from https://code.google.com/p/go-stringio/ and updated for Go 1.0.3.

### Introduction

StringIO is a low level library to mimic file I/O operations against buffers (memory). Not like File, it does not touch the filesystem at all. Best suit for testing infrastructure. Most language's standard library provides something similar to facilitate the buffer orientated IO operations, and this is an implementation for Go language.

### How to use

Its quite straightforward.

    package main
     
    import "stringio"

    func main() {
       sio := stringio.New()
       sio.WriteString("this is a test")
       sio.Seek(0, 0)
       out := sio.GetValueString()
       println(out) // prints "this is a test"
    }
 

StringIO struct implements the following interfaces:

io.ReadWriter, io.Closer, io.Seeker, io.ReadAt, io.WriteAt.

Most File I/O operations are supported, though there are a few exceptions. In most case, the are either invalid or unnecessary given there is no filesystem semantics in StringIO at all. They are:

``Fd()`` - A StringIO has no file descriptor. It doesn't need one, calling StringIO's Fd() function will return (-1, errors.New("invalid"))

``Stat()`` - StringIO has no semantics of file stats. It makes no less to implement a fake stats() call at least for now.

Dir and permission related calls, such as Chown, Chmod, Mkdir, etc... They are simply not unnecessary.

Linking/Renaming - Since StringIO has no filesystem representation, it makes no sense to implement linking and renaming at all, so no.

For more info, check out the test code. Most use cases for IO operations are tested and can be treated as examples.