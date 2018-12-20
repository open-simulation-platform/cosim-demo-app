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
- Install packr:`go get -u github.com/gobuffalo/packr/...`

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

# Build and run on Ubuntu

Tested on Ubuntu 16.04, but will most likely work on Ubuntu 18.04 and other Linux distributions as well.

Download archive and install Go (Tested with 1.11.2)  
https://golang.org/doc/install#install

Add \<install path\>/bin to PATH variable in /etc/environment.

In case you want to change the default Go workspace path (\<home path\>/go) you can set it with GOPATH in /etc/environment.

And to be able to run installed Go packages like e.g. packr without providing the full path, you might also want to add \<workspace path\>/bin to PATH variable in /etc/environment.

https://help.ubuntu.com/community/EnvironmentVariables#A.2Fetc.2Fenvironment  
Note: Variable expansion does not work in this file, and you have to log out and in again to apply changes.

Clone cse-server-go project from github.com to 
\<workspace path\>/src

Change directory to \<workspace path\>/src/cse-server-go  
Download all imported packages  
go get -d -v ./...

The next step is to pull and build the latest version of cse-core and then tell Go where to look for cse-core headers and shared libraries.
That can be done with the environment variables CGO_CFLAGS and CGO_LDFLAGS, and you can either add them permanently to e.g. /etc/environment, ~/.profile or just set them on the command line before calling go build, like shown below.

CGO_CFLAGS="-I\<cse-core path\>/include" CGO_LDFLAGS="-L\<cse-core path\>/build/output/release/lib -lcsecorec -Wl,-rpath=\<cse-core path\>/build/output/release/lib" go build

Note: It also sets the rpath (runtime search path) so that you don't have to provide it via e.g. LD_LIBRARY_PATH when you run it.

Run it  
./cse-server-go

And then open a browser at e.g. http://localhost:8000/status to verify that it's running (you should see some JSON).