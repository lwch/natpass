# build libpng
FROM debian:bullseye-slim as libpng

ARG APT_MIRROR=mirrors.ustc.edu.cn
ARG LIBPNG_VERSION=1.6.37
ARG BUILD_VERSION=0.5.0

WORKDIR /build

RUN sed -i "s|deb.debian.org|$APT_MIRROR|g" /etc/apt/sources.list && \
    sed -i "s|security.debian.org|$APT_MIRROR|g" /etc/apt/sources.list && \
    apt-get update && apt-get upgrade -y && \
    apt-get install -y gcc gcc-mingw-w64 libz-mingw-w64-dev make && \
    apt-get install -y curl && \
    curl -L http://prdownloads.sourceforge.net/libpng/libpng-$LIBPNG_VERSION.tar.gz|tar -xz && \
    cd libpng-$LIBPNG_VERSION && \
    sh -c "sed 's|^ZLIBINC.*$|ZLIBINC=/usr/x86_64-w64-mingw32/include|g' scripts/makefile.gcc | \
    sed 's|^ZLIBLIB.*$|ZLIBLIB=/usr/x86_64-w64-mingw32/lib|g' | \
    sed 's|^CC.*$|CC=x86_64-w64-mingw32-gcc|g' > Makefile" && \
    make clean && make && cp libpng.a /build/libpng.x86_64.a && \
    sh -c "sed 's|^ZLIBINC.*$|ZLIBINC=/usr/i686-w64-mingw32/include|g' scripts/makefile.gcc | \
    sed 's|^ZLIBLIB.*$|ZLIBLIB=/usr/i686-w64-mingw32/lib|g' | \
    sed 's|^CC.*$|CC=i686-w64-mingw32-gcc|g' > Makefile" && \
    make clean && make && cp libpng.a /build/libpng.i686.a

FROM debian:bullseye-slim

ARG APT_MIRROR=mirrors.ustc.edu.cn
ARG GO_VERSION=1.17.2
ARG GO_PROXY=https://goproxy.cn,direct

VOLUME /build
WORKDIR /build

RUN sed -i "s|deb.debian.org|$APT_MIRROR|g" /etc/apt/sources.list && \
   sed -i "s|security.debian.org|$APT_MIRROR|g" /etc/apt/sources.list && \
   apt-get update && apt-get upgrade -y && \
   apt-get install -y gcc libc6-dev && \
   apt-get install -y libx11-dev xorg-dev libxtst-dev libpng++-dev && \
   apt-get install -y xcb libxcb-xkb-dev x11-xkb-utils libx11-xcb-dev libxkbcommon-x11-dev && \
   apt-get install -y libxkbcommon-dev && \
   apt-get install -y xsel xclip && \
   apt-get install -y gcc-multilib make && \
   apt-get install -y gcc-mingw-w64 && \
   apt-get install -y libz-mingw-w64-dev && \
   apt-get install -y curl git && \
   curl -L https://dl.google.com/go/go$GO_VERSION.linux-amd64.tar.gz|tar -xz -C /usr/local && \
   cp -f /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

COPY --from=libpng /build/libpng.x86_64.a /usr/x86_64-w64-mingw32/lib/libpng.a
COPY --from=libpng /build/libpng.i686.a /usr/i686-w64-mingw32/lib/libpng.a

ENV PATH=$PATH:/usr/local/go/bin
ENV GOPROXY=$GO_PROXY

CMD /bin/bash