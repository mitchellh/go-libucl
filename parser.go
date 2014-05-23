package libucl

import (
	"errors"
	"sync"
	"unsafe"
)

// #include "go-libucl.h"
import "C"

// MacroFunc is the callback type for macros.
type MacroFunc func(string)

// ParserFlag are flags that can be used to initialize a parser.
//
// ParserKeyLowercase will lowercase all keys.
//
// ParserKeyZeroCopy will attempt to do a zero-copy parse if possible.
type ParserFlag int

const (
	ParserKeyLowercase ParserFlag = C.UCL_PARSER_KEY_LOWERCASE
	ParserZeroCopy                = C.UCL_PARSER_ZEROCOPY
	ParserNoTime                  = C.UCL_PARSER_NO_TIME
)

// Keeps track of all the macros internally
var macros map[int]MacroFunc = nil
var macrosIdx int = 0
var macrosLock sync.Mutex

// Parser is responsible for parsing libucl data.
type Parser struct {
	macros []int
	parser *C.struct_ucl_parser
}

// ParseString parses a string and returns the top-level object.
func ParseString(data string) (*Object, error) {
	p := NewParser(0)
	defer p.Close()
	if err := p.AddString(data); err != nil {
		return nil, err
	}

	return p.Object(), nil
}

// NewParser returns a parser
func NewParser(flags ParserFlag) *Parser {
	return &Parser{
		parser: C.ucl_parser_new(C.int(flags)),
	}
}

// AddString adds a string data to parse.
func (p *Parser) AddString(data string) error {
	cs := C.CString(data)
	defer C.free(unsafe.Pointer(cs))

	result := C.ucl_parser_add_string(p.parser, cs, C.size_t(len(data)))
	if !result {
		errstr := C.ucl_parser_get_error(p.parser)
		return errors.New(C.GoString(errstr))
	}
	return nil
}

// AddFile adds a file to parse.
func (p *Parser) AddFile(path string) error {
	cs := C.CString(path)
	defer C.free(unsafe.Pointer(cs))

	result := C.ucl_parser_add_file(p.parser, cs)
	if !result {
		errstr := C.ucl_parser_get_error(p.parser)
		return errors.New(C.GoString(errstr))
	}
	return nil
}

// Closes the parser. Once it is closed it can no longer be used. You
// should always close the parser once you're done with it to clean up
// any unused memory.
func (p *Parser) Close() {
	C.ucl_parser_free(p.parser)

	if len(p.macros) > 0 {
		macrosLock.Lock()
		defer macrosLock.Unlock()
		for _, idx := range p.macros {
			delete(macros, idx)
		}
	}
}

// Retrieves the root-level object for a configuration.
func (p *Parser) Object() *Object {
	obj := C.ucl_parser_get_object(p.parser)
	if obj == nil {
		return nil
	}

	return &Object{object: obj}
}

// RegisterMacro registers a macro that is called from the configuration.
func (p *Parser) RegisterMacro(name string, f MacroFunc) {
	// Register it globally
	macrosLock.Lock()
	if macros == nil {
		macros = make(map[int]MacroFunc)
	}
	for macros[macrosIdx] != nil {
		macrosIdx++
	}
	idx := macrosIdx
	macros[idx] = f
	macrosIdx++
	macrosLock.Unlock()

	// Register the index with our parser so we can free it
	p.macros = append(p.macros, idx)

	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	C.ucl_parser_register_macro(
		p.parser,
		cname,
		C._go_macro_handler_func(),
		C._go_macro_index(C.int(idx)))
}

//export go_macro_call
func go_macro_call(id C.int, data *C.char, n C.int) C.bool {
	macrosLock.Lock()
	f := macros[int(id)]
	macrosLock.Unlock()

	// Macro not found, return error
	if f == nil {
		return false
	}

	// Macro found, call it!
	f(C.GoStringN(data, n))
	return true
}
