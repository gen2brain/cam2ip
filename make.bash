#!/usr/bin/env bash

CHROOT="/usr/x86_64-pc-linux-gnu-static"
MINGW="/usr/i686-w64-mingw32"
RPI="/usr/armv6j-hardfloat-linux-gnueabi"
ANDROID="/opt/android-toolchain-arm7"

mkdir -p build

LIBRARY_PATH="$CHROOT/usr/lib:$CHROOT/lib" \
PKG_CONFIG_PATH="$CHROOT/usr/lib/pkgconfig" \
PKG_CONFIG_LIBDIR="$CHROOT/usr/lib/pkgconfig" \
CGO_LDFLAGS="-L$CHROOT/usr/lib -L$CHROOT/lib" \
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -v -x -o build/cam2ip.linux.amd64 -ldflags "-linkmode external -s -w" github.com/gen2brain/cam2ip


PKG_CONFIG="/usr/bin/i686-w64-mingw32-pkg-config" \
PKG_CONFIG_PATH="$MINGW/usr/lib/pkgconfig" \
PKG_CONFIG_LIBDIR="$MINGW/usr/lib/pkgconfig" \
CGO_LDFLAGS="-L$MINGW/usr/lib" \
CGO_CFLAGS="-I$MINGW/usr/include" \
CC="i686-w64-mingw32-gcc" CXX="i686-w64-mingw32-g++" \
CGO_ENABLED=1 GOOS=windows GOARCH=386 go build -v -x -o build/cam2ip.exe -ldflags "-linkmode external -s -w '-extldflags=-static'" github.com/gen2brain/cam2ip

PKG_CONFIG="/usr/bin/armv6j-hardfloat-linux-gnueabi-pkg-config" \
PKG_CONFIG_PATH="$RPI/usr/lib/pkgconfig" \
PKG_CONFIG_LIBDIR="$RPI/usr/lib/pkgconfig" \
CGO_LDFLAGS="-L$RPI/usr/lib" \
CGO_CFLAGS="-I$RPI/usr/include" \
CC="armv6j-hardfloat-linux-gnueabi-gcc" CXX="armv6j-hardfloat-linux-gnueabi-g++" \
CGO_ENABLED=1 GOOS=linux GOARCH=arm go build -v -x -o build/cam2ip.linux.arm -ldflags "-linkmode external -s -w" github.com/gen2brain/cam2ip

PATH="$PATH:$ANDROID/bin" \
PKG_CONFIG="$ANDROID/bin/arm-linux-androideabi-pkg-config" \
PKG_CONFIG_PATH="$ANDROID/lib/pkgconfig" \
PKG_CONFIG_LIBDIR="$ANDROID/lib/pkgconfig" \
CGO_LDFLAGS="-L$ANDROID/lib" \
CGO_CFLAGS="-I$ANDROID/include" \
CC="arm-linux-androideabi-gcc" CXX="arm-linux-androideabi-g++" \
CGO_ENABLED=1 GOOS=android GOARCH=arm go build -v -x -o build/cam2ip.android.arm -ldflags "-linkmode external -s -w" github.com/gen2brain/cam2ip
