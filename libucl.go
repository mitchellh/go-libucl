package libucl

// #cgo CFLAGS: -Ivendor/libucl/include
// #cgo LDFLAGS: -Lvendor/libucl/.obj -lucl
// #include <ucl.h>
import "C"

import "fmt"

func main() {
	var parser *C.struct_ucl_parser
	parser = C.ucl_parser_new(0)
	fmt.Printf("%#v\n", parser)
}
