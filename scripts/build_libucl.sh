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

# cd into the root directory
cd $DIR/..

# Create the vendor directory so we can build libucl
rm -rf vendor/libucl
mkdir -p vendor/libucl
pushd vendor/libucl
git clone https://github.com/vstakhov/libucl.git .

# Determine how to build
case $OS in
    MINGW32*)
        mingw32-make -f Makefile.w32
        cp .obj/libucl.dll .

        # Windows also needs it handy if we're testing
        cp .obj/libucl.dll ${DIR}/..
        ;;
    *)
        cmake cmake/
        make
        ;;
esac
