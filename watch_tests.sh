#!/bin/bash

autoexec -sve "GO15VENDOREXPERIMENT=1 go test ./..." main.go */*.go */*/*.go */*/*/*.go */*/*/*/*.go

