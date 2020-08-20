Cosim Demo Application
==========================
![cosim-demo-app CI](https://github.com/open-simulation-platform/cosim-demo-app/workflows/cosim-demo-app%20CI/badge.svg)

This repository contains a server-client demo application for libcosim. 
The server is written in Go and the client in clojurescript

Server
------------

### Required tools
  * Go dev tools: [Golang](https://golang.org/dl/) >= 1.11
  * Compiler: [MinGW-w64](https://sourceforge.net/projects/mingw-w64/?source=typ_redirect) (Windows), GCC >= 7 (Linux)
  * Package managers: [Conan](https://conan.io/) and [Go Modules](https://github.com/golang/go/wiki/Modules)

Throughout this guide, we will use Conan to manage C++ dependencies. However, you can also install the C++ dependencies manually.

**Note: About the installation of MinGW-w64 (Windows)**

An easier way to install it is to download the Mingw-w64 automated installer from 
[here](https://sourceforge.net/projects/mingw-w64/files/Toolchains%20targetting%20Win32/Personal%20Builds/mingw-builds/installer/mingw-w64-install.exe/download) 
and follow the steps in the wizard. It is essential that the installation path does not contain any spaces. 
Install a current version and specify win32 as thread when requested. Additionally, choose the architecture x86_64.

After installing it, you need to add it to the PATH environment variable (add the path where
your MinGW-w64 has been installed to e.g., C:\mingw\mingw64\bin). 

### Step 1: Configure Conan

First, add the OSP Conan repository as a remote and configure the username and
password to access it:

    conan remote add osp https://osp.jfrog.io/artifactory/api/conan/conan-public
    conan user -p "Open Simulation Platform" -r osp osp

### Step 2: Build and run

You can do this in two ways:

#### Alternative 1: Using Conan

From the cosim-demo-app source directory, get C/C++ dependencies using Conan:

    conan install . -u -s build_type=Release -g virtualrunenv
    go build

To run the application on Windows:

    activate_run.bat (activate_run.ps1 in PowerShell)
    cosim-demo-app.exe
    deactivate_run.bat when done (deactivate_run.ps1 in PowerShell)

To run the application on Linux:

    source activate_run.sh
    ./cosim-demo-app
    ./deactivate_run.sh when done

Open a browser at http://localhost:8000/status to verify that it's running (you should see some JSON).

#### Alternative 2: Manually handle libcosimc dependencies

You will have to define CGO environment variables with arguments pointing to your libcosimc headers and libraries. An
example for Windows can be:

    set CGO_CFLAGS=-IC:\dev\libcosimc\include
    set CGO_LDFLAGS=-LC:\dev\libcosimc\bin -lcosim -lcosimc
    go build

To run the application on Windows you need to also update the path to point to your libraries:

    set PATH=C:\dev\libcosimc\bin;%PATH%
    cosim-demo-app.exe

To run the application on Linux you need to update the LD_LIBRARY_PATH:

    LD_LIBRARY_PATH=~dev/libcosimc/lib:$LD_LIBRARY_PATH
    export LD_LIBRARY_PATH
    ./cosim-demo-app

Open a browser at http://localhost:8000/status to verify that it's running (you should see some JSON).

Client
------
Providing a web user interface.

#### Development mode
- Install a [JDK](https://www.oracle.com/technetwork/java/javase/downloads/jdk8-downloads-2133151.html)
- Install leiningen https://leiningen.org/
- Run `lein figwheel`
- View it in your browser at http://localhost:3449

You now have a framework running for live reloading of client code.

#### Building the client
- Run `lein cljsbuild once min`
- The client application will be compiled to `/resources/js/compiled`


### Create distribution with built-in client

To package the application with the client you can use packr. You can install packr and build distributable with:

    go get -u github.com/gobuffalo/packr/packr
    packr build
