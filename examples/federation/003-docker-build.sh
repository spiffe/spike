#!/usr/bin/env bash

parallel-ssh -h hosts.txt -P -t 300 "cd WORKSPACE/spike;make docker-build"

