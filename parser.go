package libucl

import (
	"errors"
	"unsafe"
)

// #include <ucl.h>
// #include <stdlib.h>
import "C"

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

// Parser is responsible for parsing libucl data.
type Parser struct {
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
}

// Retrieves the root-level object for a configuration.
func (p *Parser) Object() *Object {
	obj := C.ucl_parser_get_object(p.parser)
	if obj == nil {
		return nil
	}

	return &Object{object: obj}
}
