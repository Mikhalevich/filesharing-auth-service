#!/usr/bin/env bash

protoc -I proto/ --go_out=proto --micro_out=proto proto/auth.proto
