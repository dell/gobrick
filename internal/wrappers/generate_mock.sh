#!/bin/bash

mockgen -source wrappers.go \
  -destination wrappers_mock.go \
  -self_package github.com/dell/gobrick/internal/wrappers \
  -package wrappers
