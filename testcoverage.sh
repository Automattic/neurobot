#!/bin/bash
cd engine
go test -coverprofile=c.out && go tool cover -html=c.out
rm c.out
