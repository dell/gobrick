#!/bin/bash

mockgen -source interface.go \
  -destination multipath_mock.go \
  -package multipath
