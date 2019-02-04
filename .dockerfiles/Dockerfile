FROM ubuntu:18.04

COPY sources-18.04-azure.list /etc/apt/sources.list

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    build-essential \
    git \
    python3-setuptools \
    python3-pip \
    golang-go \
    go-dep \
    && rm -rf /var/lib/apt/lists/*

# Install  Conan package manager
RUN pip3 install conan
