# bullettime

An experimental Matrix homeserver written in Go.

Contributions are under the same terms as
https://github.com/matrix-org/synapse/blob/master/CONTRIBUTING.rst

# Starting point from Patrik Oldsberg at Ericsson:

Here's a possible starting point. What's supported so far is basic registration,
and most of the /rooms, /profile, /presence, and /events API:s.

Some explanation of the basic structure:

#### api
REST api frontend

#### db
Storage layer

#### events
Event stream implementations, only stream ordering at the moment

#### interfaces
Public interfaces that are common among all packages

#### service
Glues together lower level packages and adds business logic.

#### test
Overarching tests

#### types
Types that are common between all other packages

#### utils
General helper methods

#### vendor
external modules, uses Go 1.5 vendor experiment

To build this the repo should be cloned to $GOPATH/src/github.com/matrix-org/bullettime, then just use `go build .`
