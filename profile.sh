#!/bin/bash

TEST_NAME=TestStressTableAccess
PROFILE_NAME=cpu.prof

echo "Uncomment the test '$TEST_NAME'"
read -r -p "Press any key to continue..."

go test -cpuprofile $PROFILE_NAME -bench . -run $TEST_NAME

go tool pprof -web $PROFILE_NAME