// Copyright 2021 Frederik Zipp and others; see NOTICE file.
// Use of this source code is governed by the ISC license that
// can be found in the LICENSE file.

// Command v2go translates Verilog source code to Go.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

func usage() {
	fail(`Translates Verilog source code to Go.

Usage:
    v2go [-o go_file] [verilog_file]

    If no Verilog file is specified the tool reads from the standard input.
    If the VERILOG environment variable is set 'verilog_file' is interpreted
    relative to the path of this environment variable.

Flags:
    -o   Specifies the output file (Go). Default: standard output`)
}

func main() {
	oFlag := flag.String("o", "", "output `file` (Go)")
	flag.Usage = usage
	flag.Parse()

	var err error

	in := os.Stdin
	if flag.NArg() > 0 {
		path := flag.Arg(0)
		verilogEnv := os.Getenv("VERILOG")
		if verilogEnv != "" {
			path = filepath.Join(verilogEnv, path)
		}
		in, err = os.Open(path)
		check(err)
		defer in.Close()
	}
	check(err)

	out := os.Stdout
	if *oFlag != "" {
		out, err = os.Create(*oFlag)
		check(err)
		defer out.Close()
	}

	data, err := io.ReadAll(in)
	check(err)
	err = newParser(data, out).parseProgram()
	check(err)
}

type scanner struct {
	str     []byte
	pos     int
	nextPos int
}

var keywords = []string{
	"module", "input", "output", "reg", "wire", "assign",
	"always", "posedge", "begin", "end", "endmodule",
}

var scanRegexp = regexp.MustCompile(stripSpaces(`
      (?P<word> [_a-zA-Z][a-zA-Z0-9]* ) |
      (?P<bits> (?P<bitlen>[0-9]+) 'b (?P<bitdata>[0-1]+) ) |
      (?P<hex>  (?P<hexlen>[0-9]+) 'h (?P<hexdata>[0-9A-Fa-f]+) ) |
      (?P<int>  [0-9]+ ) |
      (?P<sym>  == | <= | [();\[:\]=^+-?{}~&|@] ) |
      \\Z`))

var skipRegexp = regexp.MustCompile("(?://.*|`.*|\\s+)*")

func newScanner(str []byte) *scanner {
	s := &scanner{str: str}
	s.skipWhitespaceAndComments(0)
	return s
}

func (s *scanner) skipWhitespaceAndComments(pos int) {
	match := skipRegexp.FindAllSubmatchIndex(s.str[pos:], 1)
	if match == nil {
		return
	}
	s.nextPos = pos + match[0][1]
}

func (s *scanner) next() (token string, value any) {
	s.pos = s.nextPos
	groups, _, end := findSubmatchGroups(scanRegexp, s.str[s.pos:])
	if groups == nil {
		panic(s.error("next token not found"))
	}
	s.skipWhitespaceAndComments(s.pos + end)
	if w, ok := groups["word"]; ok {
		for _, kw := range keywords {
			if w == kw {
				return w, ""
			}
		}
		return "<name>", w
	} else if _, ok = groups["bits"]; ok {
		bitlen, err := strconv.Atoi(groups["bitlen"])
		if err != nil {
			panic(s.error("invalid bitlen"))
		}
		bitdata, err := strconv.ParseInt(groups["bitdata"], 2, 0)
		if err != nil {
			panic(s.error("invalid bitdata"))
		}
		return "<bits>", bits{size: bitlen, value: bitdata}
	} else if _, ok = groups["hex"]; ok {
		hexlen, err := strconv.Atoi(groups["hexlen"])
		if err != nil {
			panic(s.error("invalid hexlen"))
		}
		hexdata, err := strconv.ParseInt(groups["hexdata"], 16, 0)
		if err != nil {
			panic(s.error("invalid hexdata"))
		}
		return "<bits>", bits{size: hexlen, value: hexdata}
	} else if intStr, ok := groups["int"]; ok {
		n, err := strconv.Atoi(intStr)
		if err != nil {
			panic(s.error("invalid int"))
		}
		return "<int>", n
	} else if sym, ok := groups["sym"]; ok {
		return sym, nil
	}
	return "", nil
}

func (s *scanner) where() (line, col int) {
	line = bytes.Count(s.str[:s.pos], []byte{'\n'}) + 1
	col = s.pos - bytes.LastIndexByte(s.str[:s.pos], '\n')
	return line, col
}

func (s *scanner) error(what string) scanError {
	line, col := s.where()
	return scanError{what: what, line: line, col: col}
}

type scanError struct {
	what      string
	line, col int
}

func (e scanError) Error() string {
	return fmt.Sprintf("scan errorf at line %d, column %d: %s", e.line, e.col, e.what)
}

type parser struct {
	names         map[string]bits
	registers     map[string]bits
	registerNames []string
	bitByBit      map[string][]any
	scanner       *scanner
	token         string
	value         any
	w             io.Writer
}

func newParser(text []byte, out io.Writer) *parser {
	p := &parser{
		names:     make(map[string]bits),
		registers: make(map[string]bits),
		bitByBit:  make(map[string][]any),
		scanner:   newScanner(text),
		w:         out,
	}
	p.nextToken()
	return p
}

func (p *parser) nextToken() {
	p.token, p.value = p.scanner.next()
}

func (p *parser) expected(tokens ...string) {
	line, col := p.scanner.where()
	panic(p.errorf("at line %d, column %d: expected '%s', found '%s'",
		line, col, strings.Join(tokens, " "), p.token))
}

func (p *parser) skipOver(token string) {
	if p.token != token {
		p.expected(token)
	}
	p.nextToken()
}

func (p *parser) skippingOver(tokens ...string) bool {
	for _, tok := range tokens {
		if tok == p.token {
			p.nextToken()
			return true
		}
	}
	return false
}

func (p *parser) name() string {
	n := p.value.(string)
	p.skipOver("<name>")
	return n
}

func (p *parser) int() int {
	n := p.value.(int)
	p.skipOver("<int>")
	return n
}

func (p *parser) parseProgram() (err error) {
	defer func() {
		rec := recover()
		if rec == nil {
			return
		}
		if recErr, ok := rec.(error); ok {
			err = recErr
			return
		}
		panic(rec)
	}()

	fmt.Fprint(p.w, `// Code generated by v2go; DO NOT EDIT.

package main

`)

	p.parseModule()
	for p.token != "endmodule" {
		if p.skippingOver("reg") {
			p.parseReg()
		} else if p.skippingOver("wire") {
			p.parseWire()
		} else if p.skippingOver("assign") {
			p.parseAssign()
		} else if p.skippingOver("always") {
			p.parseAlways()
		} else {
			panic(p.errorf("did not expect '%s'", p.token))
		}
	}

	fmt.Fprintln(p.w, "\nfunc eq(a, b uint64) uint64 { if a == b { return 1 }; return 0 }")

	fmt.Fprintln(p.w, "\nfunc cycle() {")
	for _, n := range p.registerNames {
		expr := p.registers[n]
		size := p.names[n].size
		fmt.Fprintf(p.w, "\t%s_tmp := (%v) & %s\n",
			n, expr.value, p.mask(size))
	}
	for _, n := range p.registerNames {
		fmt.Fprintf(p.w, "\t%s = %s_tmp\n", n, n)
	}
	fmt.Fprintln(p.w, "}")
	return nil
}

func (p *parser) getType(size int) string {
	return "uint64"
}

func (p *parser) declareInput(name string, size int) {
	typ := p.getType(size)
	p.names[name] = bits{size: size, value: name}
	fmt.Fprintf(p.w, "var %s %s\n", name, typ)
}

func (p *parser) declareWire(name string, size int) {
	p.names[name] = bits{size: size, value: name + "()"}
}

func (p *parser) declareReg(name string, size int) {
	typ := p.getType(size)
	p.names[name] = bits{size: size, value: name}
	fmt.Fprintf(p.w, "var %s %s\n", name, typ)
}

func (p *parser) parseModule() {
	p.skipOver("module")
	p.skipOver("<name>")
	p.skipOver("(")
	for !p.skippingOver("endmodule") {
		if p.skippingOver(")") {
			break
		} else if p.skippingOver("input") {
			size := p.parseDeclBitlen()
			for p.token == "<name>" {
				name := p.value.(string)
				p.declareInput(name, size)
				p.nextToken()
				if !p.skippingOver(",") {
					break
				}
			}
		} else if p.skippingOver("output") {
			size := p.parseDeclBitlen()
			for p.token == "<name>" {
				name := p.value.(string)
				p.declareWire(name, size)
				p.nextToken()
				if !p.skippingOver(",") {
					break
				}
			}
		} else {
			panic(p.errorf("don't understand module statement: %s", p.token))
		}
	}
	p.skipOver(";")
}

func (p *parser) parseDeclBitlen() int {
	if p.skippingOver("[") {
		n1 := p.int()
		p.skipOver(":")
		n2 := p.int()
		p.skipOver("]")
		if n2 != 0 {
			panic(p.errorf("end of range is not 0: %d,%d", n1, n2))
		}
		return n1 + 1
	}
	return 1
}

func (p *parser) parseWire() {
	size := p.parseDeclBitlen()
	for {
		name := p.name()
		p.declareWire(name, size)
		if p.skippingOver(";") {
			break
		}
		p.skipOver(",")
	}
}

func (p *parser) parseReg() {
	size := p.parseDeclBitlen()
	for {
		name := p.name()
		p.declareReg(name, size)
		if p.skippingOver(";") {
			break
		}
		p.skipOver(",")
	}
}

func (p *parser) parseAssign() {
	name := p.name()
	size := p.names[name].size
	if p.skippingOver("[") {
		p.parseAssignBitByBit(name)
	} else {
		p.skipOver("=")
		a := p.parseExpression()
		p.skipOver(";")
		fmt.Fprintf(p.w, "func %s() %s { return (%v) & %s }\n",
			name, p.getType(size), a.value, p.mask(size))
	}
}

func (p *parser) parseAssignBitByBit(name string) {
	size := p.names[name].size
	if _, ok := p.bitByBit[name]; !ok {
		p.bitByBit[name] = make([]any, size)
	}
	idx := p.int()
	p.skipOver("]")
	p.skipOver("=")
	a := p.parseExpression()
	if a.size != 1 {
		panic(p.errorf("expected a 1-bit expression"))
	}
	p.skipOver(";")
	p.bitByBit[name][idx] = a.value
	for _, b := range p.bitByBit[name] {
		if b == nil {
			return
		}
	}
	var parts []string
	for i := range size {
		parts = append(parts, fmt.Sprintf("((%v & 1) << %d)", p.bitByBit[name][i], i))
	}
	reverseSlice(parts)
	expr := fmt.Sprintf("(%s)", strings.Join(parts, " | "))
	fmt.Fprintf(p.w, "func %s() %s { return %s }\n", name, p.getType(size), expr)
}

func (p *parser) parseAlways() {
	p.skipOver("@")
	p.skipOver("(")
	p.skipOver("posedge")
	p.skipOver("(")
	p.skipOver("<name>")
	p.skipOver(")")
	p.skipOver(")")
	if p.skippingOver("begin") {
		for !p.skippingOver("end") {
			p.parseRegAssign()
		}
	} else {
		p.parseRegAssign()
	}
}

func (p *parser) parseRegAssign() {
	n := p.name()
	p.skipOver("<=")
	e := p.parseExpression()
	p.skipOver(";")
	p.registers[n] = e
	p.registerNames = append(p.registerNames, n)
}

func (p *parser) parseExpression() bits {
	return p.parseExprTrinary()
}

func (p *parser) parseExprTrinary() bits {
	a := p.parseExprOr()
	if !p.skippingOver("?") {
		return a
	}
	b := p.parseExprTrinary()
	p.skipOver(":")
	c := p.parseExprTrinary()
	size := p.size(b, c)
	value := fmt.Sprintf("func() uint64 { if %v > 0 { return %v }; return %v }()", a.value, b.value, c.value)
	return bits{size: size, value: value}
}

func (p *parser) parseExprOr() bits {
	a := p.parseExprXor()
	for p.skippingOver("|") {
		b := p.parseExprXor()
		size := p.size(a, b)
		value := fmt.Sprintf("(%v | %v)", a.value, b.value)
		a = bits{size: size, value: value}
	}
	return a
}

func (p *parser) parseExprXor() bits {
	a := p.parseExprAnd()
	for p.skippingOver("^") {
		b := p.parseExprAnd()
		size := p.size(a, b)
		value := fmt.Sprintf("(%v ^ %v)", a.value, b.value)
		a = bits{size: size, value: value}
	}
	return a
}

func (p *parser) parseExprAnd() bits {
	a := p.parseExprEq()
	for p.skippingOver("&") {
		b := p.parseExprEq()
		size := p.size(a, b)
		value := fmt.Sprintf("(%v & %v)", a.value, b.value)
		a = bits{size: size, value: value}
	}
	return a
}

func (p *parser) parseExprEq() bits {
	a := p.parseExprPlus()
	for p.skippingOver("==") {
		b := p.parseExprPlus()
		value := fmt.Sprintf("eq(%v, %v)", a.value, b.value)
		a = bits{size: 1, value: value}
	}
	return a
}

func (p *parser) parseExprPlus() bits {
	a := p.parseExprConcat()
	for {
		op := p.token
		if !p.skippingOver("+", "-") {
			return a
		}
		b := p.parseExprConcat()
		size := p.size(a, b)
		value := fmt.Sprintf("(%v %s %v)", a.value, op, b.value)
		a = bits{size: size, value: value}
	}
}

func (p *parser) parseExprConcat() bits {
	if !p.skippingOver("{") {
		return p.parseExprUnaryPlus()
	}
	var exprs []bits
	for {
		if p.token == "<int>" {
			exprs = append(exprs, p.parseExprRepeat())
		} else {
			exprs = append(exprs, p.parseExpression())
		}
		if !p.skippingOver(",") {
			break
		}
	}
	p.skipOver("}")
	size := 0
	var parts []string
	for i := len(exprs) - 1; i >= 0; i-- {
		b := exprs[i]
		parts = append(parts, fmt.Sprintf("((%v & %s) << %d)", b.value, p.mask(b.size), size))
		size += b.size
	}
	reverseSlice(parts)
	return bits{size: size, value: fmt.Sprintf("(%s)", strings.Join(parts, " | "))}
}

func (p *parser) parseExprRepeat() bits {
	count := p.int()
	p.skipOver("{")
	expr := p.parseExprNumber()
	p.skipOver("}")
	if expr.size != 1 {
		panic(p.errorf("can only repeat 1-bit values"))
	}
	var parts []string
	for i := range count {
		parts = append(parts, fmt.Sprintf("((%v & 1) << %d)", expr.value, i))
	}
	return bits{size: count, value: fmt.Sprintf("(%s)", strings.Join(parts, " | "))}
}

func (p *parser) parseExprUnaryPlus() bits {
	if p.skippingOver("+") || !p.skippingOver("-") {
		return p.parseExprNegation()
	}
	a := p.parseExprNumber()
	value := fmt.Sprintf("(-%v)", a.value)
	return bits{size: a.size, value: value}
}

func (p *parser) parseExprNegation() bits {
	if !p.skippingOver("~") {
		return p.parseExprNumber()
	}
	a := p.parseExprNumber()
	value := fmt.Sprintf("((^%v) & %s)", a.value, p.mask(a.size))
	return bits{size: a.size, value: value}
}

func (p *parser) parseExprNumber() bits {
	var a bits
	if p.skippingOver("(") {
		a = p.parseExpression()
		p.skipOver(")")
	} else if p.token == "<int>" {
		a = bits{value: p.value.(int), size: szNone}
		p.nextToken()
	} else if p.token == "<bits>" {
		a = p.value.(bits)
		p.nextToken()
	} else if p.token == "<name>" {
		a = p.names[p.value.(string)]
		p.nextToken()
	} else {
		panic(p.errorf("don't now message to do with %s", p.token))
	}

	if p.skippingOver("[") {
		n1 := p.int()
		var n2 int
		if p.skippingOver(":") {
			n2 = p.int()
		} else {
			n2 = n1
		}
		p.skipOver("]")
		if n1 < n2 {
			panic(p.errorf("invalid range (%d, %d)", n1, n2))
		}
		if n1 >= a.size {
			panic(p.errorf("range (%d, %d) w of bound in %s", n1, n2, a))
		}
		rsize := n1 - n2 + 1
		rvalue := fmt.Sprintf("((%v >> %d) & %s)", a.value, n2, p.mask(n1-n2+1))
		a = bits{size: rsize, value: rvalue}
	}
	return a
}

func (p *parser) mask(n int) string {
	if n == szNone {
		n = 32
	}
	return fmt.Sprintf("0x%X", uint64((1<<n)-1))
}

const szNone = -1

func (p *parser) size(a, b bits) int {
	if a.size == szNone {
		return b.size
	}
	if b.size == szNone {
		return a.size
	}
	if a.size > b.size {
		return a.size
	}
	return b.size
}

func (p *parser) errorf(format string, a ...any) parseError {
	return parseError{message: fmt.Sprintf(format, a...)}
}

type parseError struct {
	message string
}

func (e parseError) Error() string {
	return e.message
}

type bits struct {
	size  int
	value any
}

func (b bits) String() string {
	return fmt.Sprintf("<%d bits: %v>", b.size, b.value)
}

func findSubmatchGroups(re *regexp.Regexp, b []byte) (groups map[string]string, start, end int) {
	match := re.FindSubmatchIndex(b)
	if match == nil {
		return nil, 0, 0
	}
	groups = make(map[string]string, re.NumSubexp())
	for _, name := range re.SubexpNames()[1:] {
		i := re.SubexpIndex(name) * 2
		if match[i] < 0 {
			continue
		}
		groups[name] = string(b[match[i]:match[i+1]])
	}
	return groups, match[0], match[1]
}

func reverseSlice(a []string) {
	for i := len(a)/2 - 1; i >= 0; i-- {
		opp := len(a) - 1 - i
		a[i], a[opp] = a[opp], a[i]
	}
}

func stripSpaces(str string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, str)
}

func check(err error) {
	if err != nil {
		fail(err)
	}
}

func fail(msg any) {
	_, _ = fmt.Fprintln(os.Stderr, msg)
	os.Exit(1)
}
