# cout
Package cout wraps strings.Builder and adds a few useful "print into" methods to it.

     import "github.com/ohir/cout"

Cout is meant as a toolset helping to fast produce output of a PoC code and of ad-hoc cli tools:

```
                           // cout cheatsheet
                           //
  pb := cout.New(size)     // Make 'size' buffer with printers that write to it.
                           // cout.New(0) "zero buf" printers write to Stdout.
                           // cout.New(1) get buffer of MinSize size (def:256B).
        pb.AutoNL = true   // Add a nl char to print output lacking \n at end.
        pb.TrimTs = true   // Remove tail space (spaces to the newline char).
        pb.Prefix(string)  // Set a common text prefix to next writes.
                           //
  pb.Out()                 // flush to stdout (or to the 'SetOut' io.Writer).
  pb.String()              // get buffer content as string (does not copy).
  pb.SetOut(io.Writer) ok  // set where Out will flush (overide default).
                           //
                           // Printers:
  pb.Printf(fmt, ...args)  // Printf that writes to the buffer.
          p := pb.Printf   // ...often used as "p": like p("please print").
  pb.Pif(c, fmt, ...) c    // writes if c bool condition is true. Returns c.
  pb.PifNot(c, fmt, ...) c //        if c bool condition is false. Returns c.
  pb.Bar(n int, ti string) // writes "ti" titled divider, n characters wide.
  pb.NL()                  // amends buffer with an \n, if its not at the end.
  pb.ENL()                 // make sure buffer ends with an empty line.
  pb.CNL(c bool) c         // calls NL if c is true. Returns c.
  pb.CENL(c bool) c        // calls ENL if c is true. Returns c.
```
[![Go Reference](https://pkg.go.dev/badge/github.com/ohir/cout.svg)](https://pkg.go.dev/github.com/ohir/cout)

#### Knobs:
- `TrimTS` set to `true` elides all spaces at the end of lines of output (at Out time).
- `AutoNL` set to `true` adds a newline to the output of a printer method, if this output came without an ending newline.  NL is *not* added if fmt string does end with a space (for continuation prints); or if fmt ends with a newline by itself.
- Prefix, set by method `Prefix(pfx string)`, is prepended to line of output if previous fmt string did not end with a space (signalling continuation), and if current fmt string does *not* start with a newline character (signalling an intentional break).
- var `cout.MinSize` tells minimal size for non-zero buffers, eg. made with `cout.New(1)`. Default is 256B.
- var `cout.Capture` if set to non-nil io.Writer captures output of newly created cout buffers. Default is `nil`.

#### Tips:
- "Zero" buffer's printers write to *stdout* immediately. Unless `cout.Capture` variable was assigned a non-nil `io.Writer` before call to `New(0)` - then printers will write there. _Note that goodies like TrimTS and AutoNL have no effect with output going straight to the OS_.
- You can use cout methods on just declared zero buffer - ie. `var bu cout.Bld` - but until you print on it, you may not call inherited `strings.Builder` methods (with no Builder inside these will panic). If zero buffer is desired, better to obtain it via `cout.New(0)`.
- You can set `cout.Capture = os.Stderr` to change all new buffers output to stderr.
- You can capture output of all cout printers and have it layered: by eg.  `sink := cout.New(size); cout.Capture = sink` See `cout_test.go` for examples.
- Unlike a `strings.Builder`, you can copy `cout.Bld` struct (64B). But better use a pointer - as all methods are on pointer anyway.
- arguments to Bar() are optional. See package docs.

#### Caveat:
Global state (Capture, MinSize) should not be changed (used) in concurrent code. Use SetOut and explicit sizes instead.
