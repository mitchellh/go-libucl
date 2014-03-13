#!/bin/bash
#
# This script downloads and compiles libucl into a shared object.
#
set -e

# Determine our directory
SOURCE="${BASH_SOURCE[0]}"
while [ -h "$SOURCE" ] ; do SOURCE="$(readlink "$SOURCE")"; done
DIR="$( cd -P "$( dirname "$SOURCE" )" && pwd )"

# Determine the OS that we're on, which is used in some later checks.
OS=$(uname -s 2>/dev/null)

# Determine the Make information for our OS.
MAKE="make"
MAKEFILE="Makefile.unix"
case $OS in
    MINGW32*)
        MAKEFILE="Makefile.w32"
        MAKE="mingw32-make"
        ;;
    CYGWIN*)
        MAKEFILE="Makefile.w32"
        ;;
esac

# cd into the root directory
cd $DIR/..

# Create the vendor directory so we can build libucl
rm -rf vendor/libucl
mkdir -p vendor/libucl
pushd vendor/libucl
git clone https://github.com/vstakhov/libucl.git .
${MAKE} -f ${MAKEFILE}
