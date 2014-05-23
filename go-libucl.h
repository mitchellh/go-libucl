#ifndef _GOLIBUCL_H_INCLUDED
#define _GOLIBUCL_H_INCLUDED

#include <ucl.h>
#include <stdlib.h>

static inline char *_go_uchar_to_char(const unsigned char *c) {
    return (char *)c;
}

//-------------------------------------------------------------------
// Helpers: Macros
//-------------------------------------------------------------------

// This is declared in parser.go and invokes the Go function callback for
// a specific macro (specified by the ID).
extern bool go_macro_call(int, char *data, int);

// Indirection that actually calls the Go macro handler.
static inline bool _go_macro_handler(const unsigned char *data, size_t len, void* ud) {
    return go_macro_call((int)ud, (char*)data, (int)len);
}

// Returns the ucl_macro_handler that we have, since we can't get this
// type from cgo.
static inline ucl_macro_handler _go_macro_handler_func() {
    return &_go_macro_handler;
}

// This just converts an int to a void*, because Go doesn't let us do that
// and we use an int as the user data for registering macros.
static inline void *_go_macro_index(int idx) {
    return (void *)idx;
}

#endif /* _GOLIBUCL_H_INCLUDED */
