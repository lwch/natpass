ARG APT_MIRROR=mirrors.ustc.edu.cn
ARG LIBPNG_VERSION=1.6.37
ARG GO_VERSION=1.19
ARG GO_PROXY=https://goproxy.cn,direct

FROM lwch/darwin-crosscompiler:11.3

ARG APT_MIRROR
ARG GO_VERSION
ARG GO_PROXY

VOLUME /build
WORKDIR /build

RUN sed -i "s|deb.debian.org|$APT_MIRROR|g" /etc/apt/sources.list && \
   sed -i "s|security.debian.org|$APT_MIRROR|g" /etc/apt/sources.list && \
   dpkg --add-architecture i386 && \
   dpkg --add-architecture amd64 && \
   apt-get update && apt-get upgrade -y && \
   apt-get install -y gcc libc6-dev && \
   apt-get install -y gcc-multilib && \
   apt-get install -y gcc-mingw-w64 && \
   apt-get install -y curl git && \
   apt-get update && \
   apt-get install -y libx11-dev:i386 && \
   apt-get install -y libx11-dev:amd64 && \
   apt-get clean && \
   curl -L https://dl.google.com/go/go$GO_VERSION.linux-amd64.tar.gz|tar -xz -C /usr/local && \
   cp -f /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

ENV PATH=$PATH:/usr/local/go/bin
ENV GOPROXY=$GO_PROXY

CMD /bin/bash