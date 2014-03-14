package libucl

import (
	"errors"
)

// #include <ucl.h>
// #include "util.h"
import "C"

// ParserFlag are flags that can be used to initialize a parser.
//
// ParserKeyLowercase will lowercase all keys.
//
// ParserKeyZeroCopy will attempt to do a zero-copy parse if possible.
type ParserFlag int

const (
	ParserKeyLowercase ParserFlag = C.UCL_PARSER_KEY_LOWERCASE
	ParserKeyZeroCopy             = C.UCL_PARSER_ZEROCOPY
)

// Parser is responsible for parsing libucl data.
type Parser struct {
	parser *C.struct_ucl_parser
}

// ParseString parses a string and returns the top-level object.
func ParseString(data string) (*Object, error) {
	p := NewParser(0)
	defer p.Close()
	if err := p.AddChunk(data); err != nil {
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

// AddChunk adds a string chunk of data to parse.
func (p *Parser) AddChunk(data string) error {
	cstr := C.char_to_uchar(C.CString(data))
	result := C.ucl_parser_add_chunk(p.parser, cstr, C.size_t(len(data)))
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
