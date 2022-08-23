#!/bin/bash

mockgen -source interface.go \
  -destination powerpath_mock.go \
  -package powerpath
