#!/bin/bash

dir=$(pwd)
protoc --go_out=$dir *.proto
