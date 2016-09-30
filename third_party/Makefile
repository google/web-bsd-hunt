# Copyright 2016 The Web BSD Hunt Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
################################################################################
#
# TODO: High-level file comment.
include ../third_party.mk

GOPATH	:= $(shell pwd)
GOBIN	:= ${GOPATH}/bin
PATH	:= ${GOBIN}:${PATH}
export GOPATH
export GOBIN
export PATH

.PHONY: all
all: clean-govendor install-govendor install-deps

.PHONY: clean
clean: clean-govendor clean-deps

.PHONY: install
install: all

.PHONY: setup
setup:
	cd src/github.com/kardianos/govendor && go build -o ${GOBIN}/govendor

.PHONY: clean-deps
clean-deps:
	rm -rf src/*

.PHONY: install-govendor
install-govendor:
	rm -rf src/github.com/kardianos/govendor
	go get -u github.com/kardianos/govendor

.PHONY: clean-govendor
clean-govendor:
	rm -f bin/govendor

.PHONY: install-deps
install-deps: install-govendor clean-deps install-pkgs prune-deps

.PHONY: install-pkgs
install-pkgs:
	@for dep in ${THIRD_PARTY_DEPS} ; do \
		echo govendor get $$dep; \
		govendor get $$dep; \
	done

.PHONY: prune-deps
prune-deps:
	find src -type d -name .git -print0 | xargs -0 echo rm -rf
	find src -type d -name .git -print0 | xargs -0 rm -rf
