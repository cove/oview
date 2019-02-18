# Copyright © 2018 Cove Schneider
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http:#www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
VERSION = $(shell git rev-parse HEAD | cut -c 1-8)

.PHONY: build
build:
	go build  -ldflags "-X main.Version=$(VERSION)-snapshot" -o oview

.PHONY: clean
clean:
	go clean
	rm -f *.dll *.tar.gz *.zip

.PHONY: deps
deps:
	dep ensure

.PHONY: release-unix
release-unix:
	go build -ldflags "-X main.Version=$(VERSION)-release" -o oview
	tar -cvf oview-${GOOS}-${GOARCH}-$(VERSION)-release.tar oview
	gzip -9 oview-${GOOS}-${GOARCH}-$(VERSION)-release.tar

.PHONY: release-windows
release-windows:
	go build -ldflags "-X main.Version=$(VERSION})-release" -o oview
	copy vendor/github.com/g3n/engine/audio/windows/bin/*.dll .
	zip -9 oview-${GOOS}-${GOARCH}-$(VERSION)-release.tar oview *.dll
