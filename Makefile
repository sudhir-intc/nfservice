# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2019 Intel Corporation

export GO111MODULE = on

.PHONY: build build-docker nf1 nf1-docker nf2 nf2-docker clean lint help
TMP_DIR:=$(shell mktemp -d)
BUILD_DIR ?=build

VER:=1.0

build: nf1 nf2

build-docker: nf1-docker nf2-docker

nf1:
	mkdir -p "${BUILD_DIR}"
	GOOS=linux go build -o "${BUILD_DIR}/nf1" ./nf1

nf1-docker: nf1
	cp nf1/Dockerfile "${TMP_DIR}/Dockerfile_nf1"
	cp build/nf1 "${TMP_DIR}"
	cp docker-compose.yml "${TMP_DIR}"
	cp nf1/config/* "${TMP_DIR}"
	cd "${TMP_DIR}" && VER=${VER} docker-compose -f docker-compose.yml build nf1

nf2:
	mkdir -p "${BUILD_DIR}"
	GOOS=linux go build -o "${BUILD_DIR}/nf2" ./nf2

nf2-docker: nf2
	cp nf2/Dockerfile "${TMP_DIR}/Dockerfile_nf2"
	cp build/nf2 "${TMP_DIR}"
	cp docker-compose.yml "${TMP_DIR}"
	cp nf2/config/* "${TMP_DIR}"
	cd "${TMP_DIR}" && VER=${VER} docker-compose -f docker-compose.yml build nf2

clean:
	rm -rf "${BUILD_DIR}"

lint:
	golangci-lint run

help:
	@echo "Please use \`make <target>\` where <target> is one of"
	@echo "  build            to build the nf1 and nf2 test application"
	@echo "  build-docker     to build the nf1 and nf2 test application docker images"
	@echo "  nf2         to build the nf2 test application"
	@echo "  nf2-docker  to build the nf2 test application docker image"
	@echo "  nf1         to build the nf1 test application"
	@echo "  nf1-docker  to build the nf1 test application docker image"
	@echo "  clean            to clean up build artifacts"
	@echo "  lint             to run linters and static analysis on the code"
