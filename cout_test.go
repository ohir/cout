// (c) 2021 Ohir Ripe. MIT license.

/*
Testing:
  Past and current "Goldenfiles" can be written out using MKGOLD env:
  MKGOLD=FileName go test -cover # writes output to FileName for inspection.
  MKGOLD=Y make tests print to Stdout.

  There is no real "goldenfile" though - regression test uses checksums.
  Cout tests eat own's (cout's) food - by setting Capture to the common
  'sink' buffer at the first Test, then hashing its content. When test
  run individually, Capture is nil so test func prints to the stdout.

*/

package cout

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"testing"
)

var (
	commit CommitLog = func(s string) {} // stub for single tests
	aculog Bld
)

func TestAutoNew(t *testing.T) { // tests to run before Capture is set
	{
		var o Bld
		o.Clear()
		if o.sbu == nil {
			t.Logf("Expected autonew on Clear to run, but it did not")
			t.Fail()
		}
	}
	{
		var o Bld
		o.SetOut(os.Stderr)
		if o.sbu == nil {
			t.Logf("Expected autonew on SetOut to run, but it did not")
			t.Fail()
		}
		o.autonew()
	}
}

func TestFirst(t *testing.T) {
	commit, aculog, Capture = InitTestLog(1<<16, "Cout self-test")
}

func TestPrinters(t *testing.T) {
	bu := New(0) //
	p := bu.Printf
	p("This should print to stdout!\n") // or to the Capture writer
	bu.Prefix("TestBld: ")
	p("<==This should now with prefix\n<==But this line should not!\n")
	if bu.Cap() > 0 {
		t.Logf("Expected zero buffer, but got %d Cap after writes!", bu.Cap())
		t.Fail()
	}
	bu.sbu.WriteString("Misuse Zero buffer!\n")
	if bu.Cap() == 0 || bu.size != 0 {
		t.Logf("Misused buffer still has Cap %d or bad size %d!", bu.Cap(), bu.size)
		t.Fail()
	}
	bu.Out()
	commit("zerobuf")

	p("-- This print is after \"zerobuf\" commit --\n")
	bu = New(1) // now check us buffered
	m := "This should print to our buffer!\n"
	p(m)
	if bu.Len() != len(m) { // check if its at ours
		t.Logf("Expected %d in own buffer, but got %d!", len(m), bu.Len())
		t.Fail()
	}
	bu.NL()
	bu.ENL()
	bu.CNL(true)
	bu.Prefix("Pfx2: ")
	p("This should print out with prefix!\n\n")
	if bu.Len() == 0 {
		t.Logf("Expected to have filled buffer, but it has Len 0")
		t.Fail()
	}
	bu.Out()
	if bu.Len() != 0 {
		t.Logf("Expected to flush buffer, but still Len is %d!", bu.Len())
		t.Fail()
	}
	commit("rudimentary")
	bu.Prefix("")
	var inv io.Writer
	if ok := bu.SetOut(inv); ok {
		t.Logf("SetOut should have NOT set invalid Writer, but it did!")
		t.Fail()
	}
	bu.Clear()
	x, y := 3, 4
	bu.Pif(x > y, "Not printed!")
	bu.PifNot(x < y, "Not printed!")
	if bu.Len() != 0 {
		t.Logf("Pif and PifNot should'nt have printed, but did!")
		t.Fail()
	}
	bu.Pif(x < y, "Pif printed!\n")
	bu.PifNot(x > y, "PifNot printed!\n")
	if bu.Len() != 29 {
		t.Logf("Pif and PifNot should have printed, but did not! (%d)", bu.Len())
		t.Fail()
	}
	bu.Out()
	commit("PifPaf")
}

func TestNLprinters(t *testing.T) {
	bu := New(1)
	bu.Printf("[1] Testing NL printers")
	bu.Out()
	bu.NL()
	bu.ENL()
	if bu.Len() != 0 {
		t.Logf("Neither NL nor ENL should have printed to empty buffer")
		t.Fail()
	}
	bu.Printf("\n")
	bu.NL()
	if bu.String() != "\n" {
		t.Logf("NL should NOT print on an NL. >%q<", bu.String())
		t.Fail()
	}
	bu.Printf(".")
	bu.NL()
	if bu.String() != "\n.\n" {
		t.Logf("NL should print but it did not: >%q<", bu.String())
		t.Fail()
	}
	bu.ENL()
	if bu.String() != "\n.\n\n" {
		t.Logf("ENL should print single NL but it did not: >%q<", bu.String())
		t.Fail()
	}
	bu.Printf("[2] Continue testing NL printers")
	bu.Out()
	bu.Printf(".")
	bu.ENL()
	if bu.String() != ".\n\n" {
		t.Logf("ENL should print two NLs but it did not: >%q<", bu.String())
		t.Fail()
	}
	bu.Printf("[3] Continue testing NL printers")
	bu.Out()
	bu.Printf("\n")
	bu.ENL()
	if bu.String() != "\n\n" {
		t.Logf("ENL should print two NLs but it did not: >%q<", bu.String())
		t.Fail()
	}
	bu.Printf("[4] Continue testing NL printers")
	bu.Out()
	bu.Printf("..")
	bu.ENL()
	if bu.String() != "..\n\n" {
		t.Logf("ENL should print two NLs but it did not: >%q<", bu.String())
		t.Fail()
	}
	bu.ENL()
	if bu.String() != "..\n\n" {
		t.Logf("ENL should NOT print if two NLs are present, but it did: >%q<", bu.String())
		t.Fail()
	}
	{
		ob := New(0) // should be redirected in test suite
		ob.ENL()     // change Capture output by two NLs
	}
	{
		var ob Bld // zero and nil
		ob.ENL()
		if ob.sbu == nil {
			t.Logf("ENL should call autonew, but it did not")
			t.Fail()
		}
	}
	bu.Printf("[5] End NL printers test\n")
	bu.Out()
	commit("")
}

func TestAutoNL(t *testing.T) {
	bu := New(1)
	p := bu.Printf
	bu.AutoNL = true // turn on autonl
	for i, s := 2, "-- now all joined: "; i > 0; i-- {
		p("[1] This should be a separate line [1]")
		p("[2] This should be a separate line too")
		p("[3] This is leading part of the ")
		p("whole line [3].")
		p(s)
		s = "\n"
		bu.AutoNL = false // turn off autonl
	}
	nnl := strings.Count(bu.String(), "\n")
	if nnl != 4 {
		t.Logf("Expected 4 newlines in output, but got %d!", nnl)
		t.Fail()
	}
	bu.Out()
	commit("")
}

func TestPrefix(t *testing.T) {
	bu := New(1)
	p := bu.Printf
	bu.Prefix("Anl: ")
	bu.AutoNL = true // turn on autonl
	for i, s := 2, "-- now all joined: "; i > 0; i-- {
		p("[1] This should be a separate line [1] with prefix")
		p("[2] This should be a separate line too")
		p("[3] This is leading part of the ")
		p("whole line [3].")
		p("\n!!! But this line should have no prefix as it starts with \\n!")
		bu.Prefix("NoAnl: ")
		p(s)
		s = "\n"
		bu.AutoNL = false // turn off autonl
	}
	nnl := strings.Count(bu.String(), "\n")
	if nnl != 7 {
		t.Logf("Expected 7 newlines in output, but got %d!", nnl)
		t.Fail()
	}
	bu.Out()
	commit("")
}

type In2ExpStr struct{ Inp, Exp string }

func TestTrim(t *testing.T) {
	log := New(1)
	ttab := []In2ExpStr{
		{"", ""},
		{"  \nabcdef  ", "\nabcdef"},
		{" \n", "\n"},
		{" \n ", "\n"},
		{" \n  ", "\n"},
		{"  \n  ", "\n"},
		{" \n \n", "\n\n"},
		{"  \n \n", "\n\n"},
		{"  \n \n ", "\n\n"},
		{"  \n  \n ", "\n\n"},
		{"  \n  \n  ", "\n\n"},
		{"a\nb \nc  ", "a\nb\nc"},
		{"\n b\n a ", "\n b\n a"},
		{"  \n \n  ", "\n\n"},
		{"  \n\n", "\n\n"},
		{" ", ""},
		{
			"Some lines   \n  \nhave tails \nof space\nthat clobber Examples \n \n \n",
			"Some lines\n\nhave tails\nof space\nthat clobber Examples\n\n\n",
		},
	}
	o2cmp := New(1)
	o2test := New(1)
	o2test.TrimTs = true
	o2test.SetOut(o2cmp) // redirect Out to our compare
	for i, ti := range ttab {
		mInp, mExp := ti.Inp, ti.Exp
		if len(mInp) > 0 && mInp[0] == '#' {
			break
		}
		o2test.Printf(mInp)
		o2test.Out() // trimmed to o2cmp
		if o2cmp.String() != mExp {
			em := fmt.Sprintf("tab[%d] Expected: %q\ntab[%d] but it does not match: %q!",
				i, mExp, i, o2cmp.String())
			log.WriteString(em)
			t.Log(em)
		}
		o2test.Clear()
		o2cmp.Clear()
	}
	if log.Len() > 0 { // failed
		t.Fail()
		log.WriteString("\nTESTS FAILED!\n")
		log.Out() // to stdout or capture
		commit("FAILED!")
	} else {
		log.WriteString("Tail clean procedures tested OK!\n")
		log.Out() // to stdout or capture
		commit("OK!")
	}
}

func TestTrimANL(t *testing.T) {
	log := New(1)
	o2cmp := New(1)
	o2test := New(1)
	o2test.TrimTs = true
	o2test.AutoNL = true
	o2test.SetOut(o2cmp) // redirect Out to our compare
	nltab := []In2ExpStr{
		{"", ""},
		{" ", "\n"},    // corner
		{"    ", "\n"}, // corner
		{"     \n  ", "\n"},
		{"", ""},
		{"  \nabab", "\nabab\n"},
		{"  \nabab   \n   ", "\nabab\n"},
		{"     \n  ", "\n"},
		{"", ""},
	}
	for i, ti := range nltab {
		mInp, mExp := ti.Inp, ti.Exp
		if len(mInp) > 0 && mInp[0] == '#' {
			break
		}
		o2test.Printf(mInp)
		o2test.Out() // trimmed to o2cmp
		if o2cmp.String() != mExp {
			em := fmt.Sprintf("tab[%d] Expected: %q\ntab[%d] but it does not match: %q!",
				i, mExp, i, o2cmp.String())
			log.WriteString(em)
			t.Log(em)
		}
		o2test.Clear()
		o2cmp.Clear()
	}
	if log.Len() > 0 { // failed
		t.Fail()
		log.WriteString("\nANL TESTS FAILED!\n")
		log.Out() // to stdout or capture
		commit("FAILED!")
	} else {
		log.WriteString("ANL tail clean procedures tested OK!\n")
		log.Out() // to stdout or capture
		commit("OK!")
	}
}

func TestBar(t *testing.T) {
	bu := New(1) //
	// p := bu.Printf
	bu.Bar()
	bu.Bar(40)
	bu.Bar(60, "~~~ sixty tildes ")
	bu.Prefix("Pfx: ")
	bu.Bar()
	bu.Bar(40)
	bu.Bar(60, "~~~ prefixed sixty tildes ")
	bu.Bar(0, "~~~ do not print ")
	bu.Bar(10, "~~~ *ten* prefixed tildes ")
	bu.Bar(30, "")
	exp := 427
	if bu.Len() != exp {
		t.Logf("Expected %d characters in buffer, but got %d instead!", exp, bu.Len())
		t.Fail()
	}
	bu.Out()
	commit("")
}

func TestAutonew(t *testing.T) {
	{
		var x Bld
		x.Printf("Hello! From uninitialized Bld...")
		x.Out()
	}
	{
		var x Bld
		x.NL()
		x.Out()
	}
	{
		var x Bld
		x.CNL(true)
		x.CENL(true)
		x.Printf("..Line above should be empty\n")
		x.Out()
	}
	commit("")
}

/*func TestX(t *testing.T) {
	bu := New(1) //
	p := bu.Printf
	exp := 1
	if bu.Len() != exp {
		t.Logf("Expected %d characters in buffer, but got %d instead!", exp, bu.Len())
		t.Fail()
	}
	bu.Out()
	commit("")
}*/

func TestLast(t *testing.T) {
	all := aculog.String()
	skipck := false
	if outfn := os.Getenv("MKGOLD"); outfn != "" {
		if outfn == "F" {
			aculog.Out()
		} else if outfn == "NOHASH" {
			skipck = true
		} else if outfn == "Y" {
			aculog.Out()
		} else if err := os.WriteFile(outfn, []byte(all), 0660); err != nil {
			t.Fatalf("Can not dump to file %s [%v]", outfn, err)
		} else {
			fmt.Fprintf(os.Stderr, "Output has been written to %s\n", outfn)
		}
	}
	if got, ok := IsCksumOK(all, expectedDjb); !ok && !skipck {
		t.Logf("\nChecksums do not match!\n"+
			"Registered sum is  %#x\n"+
			"    Now I have got %#x\n"+
			"    Do INSPECT WHY before correcting sum in cout_test.go to be:\n\n"+
			"const expectedDjb = %#[2]x\n\n"+
			"    Dump previous and current output to files using:\n"+
			"    MKGOLD=filename go test\n",
			uint64(expectedDjb), got)
		t.Fail()
	}
}

/*
TODO(ohir) make "Capture" tests public for reuse. Not yet done.

func CommitLog type is used during Go tests to register subsequent
 tests (after the test). See cout_test.go for example of usage.

Ie: in *first to run* of the the _test.go files declare the package globals:
	var commit cout.CommitLog = func(s string) {} // stub for single tests
	var aculog cout.Bld

Then in first to run test function assign to them:

func TestFirst(t *testing.T) {
	commit, aculog, cout.Capture = InitTestLog(1<<16, "Cout self-test")
}
*/

// Testing helper IsCksumOK tests whether 'have' content hashes to
// 'expect' checksum - if it does, 'ok' is true. With djbnz default,
// hash is zero if input contains even a single zero byte.
func IsCksumOK(have string, expect uint64) (hh uint64, ok bool) {
	hh = cksum(have)
	return hh, hh == expect
}

type CommitLog func(desc string)

// Normally tests of cout Printer will print to Stdout if ran independently.
// Full suite will capture output to acu and checksum subresults by calling:
// commit, aculog, Capture = InitTestLog(size, "title" )  in the first Test.
func InitTestLog(size int, lead string) (logf CommitLog, acu, catch Bld) {
	const tWidth = 99
	acu, catch = New(size), New(size/2)
	if len(lead) == 0 {
		lead = "cout.CommitLog init"
	}
	if len(lead) < tWidth-6 {
		bar := strings.Repeat("-", (tWidth-len(lead))/2)
		acu.Printf("%s %s %s\n", bar, lead, bar)
	} else {
		acu.Printf("%s\n", lead)
	}
	return func(desc string) {
		who := forwhom(3)
		acu.Printf("\n %s:%s output >>> ", who, desc)
		rs := catch.String()
		rh := cksum(rs)
		if rh == 0 {
			acu.WriteString(" CONTAINS ZERO CHARACTER! ")
		}
		acu.Printf("subhash: %016x >>>\n\n%s", rh, rs)
		catch.Clear()
	}, acu, catch
}

// func forwhom returns the name of function who called it up the chain.
// 'up' gives the position of sought stack frame, with current being #1.
// Ie. 'up' 2 means caller of forwhom, 3 caller of caller of forwhom...
// Name not always can be estabilished, eg. for frame of inlined code.
func forwhom(up int) (r string) {
	fra := []uintptr{0}
	const unknown = "UnknownFunc"
	if n := runtime.Callers(up, fra); n != 1 {
		return unknown
	}
	frames := runtime.CallersFrames(fra)
	yfr, _ := frames.Next()
	var fna string
	if fna = yfr.Function; len(fna) == 0 {
		return unknown
	}
	if dp := strings.LastIndexByte(fna, '.'); dp >= 0 && dp < len(fna)-1 {
		fna = fna[dp+1:]
	} // return fmt.Sprintf("%d:%s", yfr.Line-1, fna) no lines!
	return fna
}

var cksum func(string) uint64 = djbnz

func djbnz(in string) (rh uint64) { // keep sum version
	rh = 5681
	for i, c := 0, byte(0); i < len(in); i++ {
		if c = in[i]; c == 0 { // non-zero check
			return 0
		}
		rh = ((rh<<5 + rh) + uint64(c))
	}
	// As 33 does not share any common divisors with 2^64, assumptions
	// of original (sum) Djb hash stand firm also for the 64b version.
	// Colliding input can be constructed, of course, but not without
	// effort. When you get legitimate one, open issue at github.
	// There is commented out sha1 based regression test in
	// cout_test.go. Uncomment and use it if in doubts.
	return rh
}

// All tests output "goldenhash"
const expectedDjb = 0x8e197c26d02604ff
