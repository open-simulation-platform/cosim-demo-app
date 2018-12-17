# cse-client
Providing graphical user interface for CSE.

## Development mode
- Install a JDK https://www.oracle.com/technetwork/java/javase/downloads/jdk8-downloads-2133151.html
- Install leiningen https://leiningen.org/
- Run `lein figwheel`
- View it in your browser at http://localhost:3449

You now have a framework running for live reloading of client code.

### Server
The client connects to the cse-server on http://localhost:8000, run it separately. Do not open this one in your browser when developing.

# Building the client
- Run `lein cljsbuild once min`
- The client application will be compiled to `/resources/js/compiled`
- copy the resource folder to the cse-server root to use the built client