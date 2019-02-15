FROM ubuntu:18.04

COPY sources-18.04-azure.list /etc/apt/sources.list

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    build-essential \
    patchelf \
    git \
    python3-setuptools \
    python3-wheel \
    python3-pip \
    golang-go \
    go-dep \
    && rm -rf /var/lib/apt/lists/*

# Install packr
ENV GOPATH=/go
ENV PATH=${GOPATH}/bin:${PATH}
RUN go get -v -u github.com/gobuffalo/packr/packr

# Install  Conan package manager
RUN pip3 install conan
