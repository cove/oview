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

.PHONY: build
build:
	go build -o oq

.PHONY: clean
clean:
	go clean
	rm -f *.dll *.tar.gz *.zip

.PHONY: deps
deps:
	dep ensure

.PHONY: release-unix
release-unix:
	go build -o oq
	tar -cvf oq-${GOOS}-${GOARCH}.tar oq
	gzip -9 oq-${GOOS}-${GOARCH}.tar

.PHONY: release-windows
release-windows:
	go build -o oq
	copy vendor/github.com/g3n/engine/audio/windows/bin/*.dll .
	zip -9 oq-${GOOS}-${GOARCH}.tar oq *.dll

