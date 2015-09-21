# bullettime

An experimental Matrix homeserver written in Go.

Contributions are under the same terms as
https://github.com/matrix-org/synapse/blob/master/CONTRIBUTING.rst

# Starting point from Patrik Oldsberg at Ericsson:

Here's a possible starting point. What's supported so far is basic registration,
and most of the /rooms, /profile, /presence, and /events API:s.

To build this the repo should be cloned to $GOPATH/src/github.com/matrix-org/bullettime, then just use `go build .`

Some explanation of the basic structure:

- #### core/
Core functionality that uses the data structures in the Matrix spec
    - **db/**
    Storage abstractions

    - **events/**
    Event stream implementations, only stream ordering at the moment

    - **interfaces/**
    Public interfaces that are common among all packages

    - **types/**
    Types that are common among all other packages

- #### matrix/
Implementation of the business logic in the Matrix spec, on top of core
    - **api/**
    REST api frontend, built on top of services

    - **service/**
    Glues together core packages and adds business logic.

    - **interfaces/**
    Public interfaces that are common among matrix packages

    - **types/**
    Types that are common among matrix packages

- #### test/
Overarching tests

- #### utils/
General helper methods

- #### vendor/
external modules, uses Go 1.5 vendor experiment
