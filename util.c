// Utility function to convert a char* to unsigned char* because
// the libucl API takes this and we can't do this in pure Go.
unsigned char *char_to_uchar(char *original) {
    return (unsigned char *)original;
}
