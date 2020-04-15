#!/bin/bash

mockgen -source interface.go \
  -destination scsi_mock.go \
  -package scsi
