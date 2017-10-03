#!/usr/bin/env bash

CHROOT="/usr/x86_64-pc-linux-gnu-static"
MINGW="/usr/i686-w64-mingw32"
RPI="/usr/armv6j-hardfloat-linux-gnueabi"

mkdir -p build

LIBRARY_PATH="$CHROOT/usr/lib:$CHROOT/lib" \
PKG_CONFIG_PATH="$CHROOT/usr/lib/pkgconfig" \
PKG_CONFIG_LIBDIR="$CHROOT/usr/lib/pkgconfig" \
CGO_LDFLAGS="-L$CHROOT/usr/lib -L$CHROOT/lib" \
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -v -x -o build/cam2ip.amd64 -ldflags "-linkmode external -s -w" github.com/gen2brain/cam2ip


PKG_CONFIG="/usr/bin/i686-w64-mingw32-pkg-config" \
PKG_CONFIG_PATH="$MINGW/usr/lib/pkgconfig" \
PKG_CONFIG_LIBDIR="$MINGW/usr/lib/pkgconfig" \
CGO_LDFLAGS="-L$MINGW/usr/lib -L$MINGW/usr/include" \
CGO_CFLAGS="-I$MINGW/usr/include" \
CC="i686-w64-mingw32-gcc" CXX="i686-w64-mingw32-g++" \
CGO_ENABLED=1 GOOS=windows GOARCH=386 go build -v -x -o build/cam2ip.exe -ldflags "-linkmode external -s -w '-extldflags=-static'" github.com/gen2brain/cam2ip

PKG_CONFIG="/usr/bin/armv6j-hardfloat-linux-gnueabi-pkg-config" \
PKG_CONFIG_PATH="$RPI/usr/lib/pkgconfig" \
PKG_CONFIG_LIBDIR="$RPI/usr/lib/pkgconfig" \
CGO_LDFLAGS="-L$RPI/usr/lib -L$RPI/usr/include" \
CGO_CFLAGS="-I$RPI/usr/include" \
CC="armv6j-hardfloat-linux-gnueabi-gcc" CXX="armv6j-hardfloat-linux-gnueabi-g++" \
CGO_ENABLED=1 GOOS=linux GOARCH=arm go build -v -x -o build/cam2ip.arm -ldflags "-linkmode external -s -w" github.com/gen2brain/cam2ip
