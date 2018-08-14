# cse-server-go
playground for testing go as an alternative for the cse-server

# Get going (windows)
- install [go](https://golang.org/dl/)
- environmental variables must be defined
    - [GOROOT](https://golang.org/doc/install#tarball_non_standard) - directory where go is installed
    - [GOPATH](https://golang.org/cmd/go/#hdr-GOPATH_environment_variable) - where to look for Go files
- install [MinGW-w64](https://sourceforge.net/projects/mingw-w64/?source=typ_redirect) (TODO: Should investigate using VS)
- csecorec.dll and csecorecpp.dll must be manually copied into cse-server-go root folder
- go build will compile executable

# Dependencies
- Install the dep tool https://golang.github.io/dep/
- Type `dep ensure` in your shell to download deps to your disk.

# Interactive web development (client)
- Install https://leiningen.org/ (and possibly a Java JDK)
- cd /client
- lein figwheel

# Interactive web development (server)
- Install https://github.com/codegangsta/gin
- Did not get any further on this, but should be the way to go.

# Building binary distribution
- In client folder: Run `lein cljsbuild once min`
- In server root folder: Run `packr build`