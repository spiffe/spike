#!/usr/bin/env bash

parallel-ssh -h hosts.txt -P "cd WORKSPACE/spike;make k8s-start"

