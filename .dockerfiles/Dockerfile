FROM ubuntu:20.04

COPY sources-20.04-azure.list /etc/apt/sources.list

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    build-essential \
    curl \
    patchelf \
    git \
    flex \
    python3-setuptools \
    python3-wheel \
    python3-pip \
    && rm -rf /var/lib/apt/lists/*

# Install Go
RUN curl "https://dl.google.com/go/go1.11.5.linux-amd64.tar.gz" | tar xvz -C /usr/local; \
    export PATH="/usr/local/go/bin:$PATH"; \
    go version

ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH

RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 777 "$GOPATH"
WORKDIR $GOPATH

# Install additional build tools
RUN go get -v -u github.com/gobuffalo/packr/packr && go clean -cache
RUN pip3 install conan cmake
