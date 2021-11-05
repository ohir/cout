// (c) 2021 Ohir Ripe. MIT license.

/* Package cout wraps strings.Builder and adds a few useful "print into" methods to it.
Cout is meant as a toolbox helping us to fast write a PoC code and ad-hoc cli tools.

  import "github.com/ohir/cout"

                         cout cheatsheet
  pb := cout.New(size)     // Make 'size' buffer with printers that write to it.
                           // cout.New(0) "zero buf" printers write to Stdout.
                           // cout.New(1) get buffer of MinSize size (def:256B).
        pb.AutoNL = true   // Add a nl char to format strings lacking \n at end.
        pb.TrimTs = true   // Remove tail space (spaces to the newline char).
        pb.Prefix(string)  // Set a common text prefix to all next writes.
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
*/
package cout

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// Package config: MinSize of created buffer, and Capture io.Writer.
var (
	Capture io.Writer // used instead of Stdout, if set
	MinSize = 1 << 8  // of buffer
)

// type Bld exposes cout API. It exposes also TrimTS, AutoNL, and Prefix(string) knobs.
type (
	Bld struct { // use cout.New
		*sbu             // our strings.Builder
		size   int       // initial Builder size
		AutoNL bool      // add newline unless fmt ends w/space or NL
		TrimTs bool      // trim tailspace at Out() calling time
		haspfx bool      // prefix on/off
		skipfx bool      // skip prefix (call to call)
		pfx    []byte    // Prefix with this if not in chain
		to     io.Writer // printers write to
		wout   io.Writer // Out() flushes to.
	}
	sbu = strings.Builder
)

// see fmt.Printf docs
func (b *Bld) Printf(fm string, a ...interface{}) {
	end := len(fm) - 1
	switch {
	case end < 0:
		return
	case b.sbu == nil:
		b.autonew()
	}
	if b.haspfx && !b.skipfx && fm[0] != '\n' {
		b.to.Write(b.pfx)
	}
	fmt.Fprintf(b.to, fm, a...)
	if b.AutoNL && fm[end] != '\n' && fm[end] != ' ' {
		b.NL()
	}
	b.skipfx = fm[end] == ' '
}

// func cout.New returns wrapped strings.Builder of requested size
// with added simple printers: Printf, Bar, NL, ENL - and their conditional
// variants.  On returned Bld struct a complete strings.Builder API can be
// used too.
//
// Zero buffer's printers print directly to the output io.Writer, which
// defaults to os.Stdout - unless cout.Capture var was assigned a non-nil
// io.Writer before a call to New.  Non zero buffer's printers write into
// internal buffer, then accumulated content can be retrieved anytime
// using String() method, or it can be flushed once, with method Out().
//
//    logstd := cout.NewBuf(0) // zero buffer - writes to Stdout direct
//    logbuf := cout.NewBuf(1) // default size buffer, Out() to Stdout
//    logbuf.SetOut(os.Stderr) // now Out() will flush to Stderr
//
func New(size int) Bld {
	var cf Bld
	var sb sbu
	cf.sbu = &sb
	cf.wout = Capture
	if cf.wout == nil {
		cf.wout = os.Stdout
	}
	if size == 0 {
		cf.to = cf.wout
	} else {
		cf.to = &sb
		cf.size = MinSize
		if size > MinSize {
			cf.size = size
		}
		cf.Grow(cf.size)
	}
	return cf
}

func (b *Bld) autonew() {
	if b.sbu != nil {
		return
	}
	var sb sbu
	b.sbu = &sb
	b.wout = Capture
	if b.wout == nil {
		b.wout = os.Stdout
	}
	b.to = b.wout
}

// Method SetOut allows to change where method Out() will dump buffer
// content. Default output is set to Stdout, unless cout.Capture var
// was assigned a non-nil io.Writer before cout.NewBuf(size) call.
func (b *Bld) SetOut(w io.Writer) (ok bool) {
	switch {
	case w == nil:
		return
	case b.sbu == nil:
		b.autonew()
	}
	b.wout = nil
	b.wout = w
	return true
}

// Method Out flushes buffer to the output Writer (ie. Capture, then Stdout)
// then it calls Clear()
func (b *Bld) Out() {
	if b.sbu == nil || b.Cap() == 0 || b.Len() == 0 {
		return
	}
	if !b.TrimTs {
		fmt.Fprint(b.wout, b.String())
		b.Clear()
		return
	} // else trim all tails
	s, tol := b.String(), 0
	for {
		if at := strings.Index(s, " \n"); at >= 0 {
			for tol = at + 1; tol > 0 && s[tol-1] == ' '; tol-- {
			}
			fmt.Fprint(b.wout, s[:tol])
			s = s[at+1:]
			continue
		}
		for tol = len(s); tol > 0 && s[tol-1] == ' '; tol-- {
		}
		fmt.Fprint(b.wout, s[:tol])
		switch {
		case !b.AutoNL:
		case tol < 1, s[tol-1] != '\n':
			b.wout.Write([]byte("\n"))
		}
		break
	}
	b.Clear()
}

// Method Clear removes content and sets buffer to its initial size
// It does not touch other settings. Prefer Clear to Reset.
func (b *Bld) Clear() {
	switch {
	case b.sbu == nil:
		b.autonew()
	case b.size > 0 && b.Len() > 0:
		b.Reset()
		b.Grow(b.size)
	case b.Cap() > 0: // zero builder used as sink
		b.Reset()
	}
}

// Method Prefix sets text to be prepended at whole output lines.
// Set prefix does not print if current fmt string starts with a newline,
// or previous fmt string ended with space.
func (b *Bld) Prefix(pfx string) { b.pfx = []byte(pfx); b.haspfx = len(pfx) > 0 }

// Method Pif writes if the c condition is true. Returns c as given.
func (b *Bld) Pif(c bool, fm string, a ...interface{}) bool {
	if !c {
		return c
	}
	b.Printf(fm, a...)
	return c
}

// Method PifNot writes if the c condition is false. Returns c as given.
func (b *Bld) PifNot(c bool, fm string, a ...interface{}) bool {
	if c {
		return c
	}
	b.Printf(fm, a...)
	return c
}

// Method NL amends buffer with a single newline if buffer is not empty
// and if it does not already has a NL character at end. For zero buffers
// or redirected output single NL is written always, as we can
// inspect only our own buffer.
func (b *Bld) NL() {
	// somehow Pif/PifNot mostly dealt with '\n',
	// make such usecases a single call
	switch {
	case b.sbu == nil:
		b.autonew()
		fallthrough
	case b.to != b.sbu:
		b.to.Write([]byte{'\n'})
		b.skipfx = false
	case b.Len() > 0 && b.String()[b.Len()-1] != '\n':
		b.to.Write([]byte{'\n'})
		b.skipfx = false
	}
}

// Method ENL amends non-zero, not empty buffer in a way that there will
// be two NL chars at the end. Ie. it makes sure an empty line will print.
// For zero buffers or redirected output it writes NLNL unconditionally.
func (b *Bld) ENL() {
	const nl byte = '\n'
	const nlnl string = "\n\n"
	var b2, b1 byte
	switch {
	case b.sbu == nil:
		b.autonew()
		fallthrough
	case b.to != b.sbu:
		b.to.Write([]byte(nlnl))
		b.skipfx = false
		return
	case b.Len() == 0:
		return
	case b.Len() == 1:
		b1 = b.String()[0]
	default:
		b2, b1 = b.String()[b.Len()-2], b.String()[b.Len()-1]
	}
	switch {
	case b2 != nl && b1 == nl:
		b.to.Write([]byte{nl})
		b.skipfx = false
	case b2 != nl && b1 != nl:
		b.to.Write([]byte(nlnl))
		b.skipfx = false
	}
}

// func CNL conditionally calls NL(); then returns condition intact.
func (b *Bld) CNL(c bool) bool {
	if c {
		b.NL()
	}
	return c
}

// func CENL conditionally calls ENL(); then returns condition intact.
func (b *Bld) CENL(c bool) bool {
	if c {
		b.ENL()
	}
	return c
}

// Method Bar() writes 79 dashes as output divider. If called with one or two
// optional parameters: Bar(width int, title string), it writes title followed
// by repeated first character of the title. Eg. usage: `p.Bar(77,"~~ tildes ")`
func (b *Bld) Bar(a ...interface{}) { // title, width
	blen := 79
	bstr := "-"
	for _, aa := range a {
		switch v := aa.(type) {
		case string:
			bstr = v
		case int:
			blen = v
			if blen < 1 { // user said off
				return
			}
		}
	}
	if len(bstr) == 0 {
		bstr = "="
	}
	tail := blen - len(bstr) - len(b.pfx)
	if tail < 0 {
		tail = 0
	}
	// can be prefixed
	b.Printf("%s%s\n", bstr, strings.Repeat(string(bstr[0]), tail))
}
