#!/usr/bin/env bash

# Source the common.sh script
# shellcheck source=./common.sh
. "$(git rev-parse --show-toplevel || echo ".")/scripts/common.sh"

cd "$PROJECT_DIR" || exit 1

# Installing tools which are needed
#-------------------------------------------------------------------------------
# install spkit first
if [ ! -f "$GOPATH/bin/spkit" ]; then
  echo_info "Install spkit"
  mkdir -p "$GOPATH/bin"
  wget "https://spkit.shopee.io/spkit/stable/spkit-$(uname -s| tr '[:upper:]' '[:lower:]')" -O "$GOPATH/bin/spkit"
  chmod a+x "$GOPATH/bin/spkit"
fi

# install common tools via spkit cmd
go mod tidy
spkit install --debug

# install mockgen
if ! has mockgen; then
  echo_info "Install mockgen"
  spkit run go get -v github.com/golang/mock/gomock
  spkit run go install -v -i github.com/golang/mock/mockgen
fi

# install easytags, which adds json or yaml tags into golang struct.
if ! has easytags; then
  echo_info "Install easytags (generate and update struct tags)"
  spkit run go get github.com/betacraft/easytags
fi

# install protoc-gen-gofast
if ! has protoc-gen-gofast; then
  echo_info "Install protoc-gen-gofast"
  go get github.com/gogo/protobuf/protoc-gen-gofast
  if ! has protoc-gen-gofast; then
    echo_error "Please Try go 1.14 to install protoc-gen-gofast if failed"
  fi
fi

# install protoc-gen-gogo-spex-rpc
if ! has protoc-gen-gogo-spex-rpc; then
  echo_info "Install protoc-gen-gogo-spex-rpc"
  go get git.garena.com/shopee/common/spkit/cmd/protoc-gen-gogo-spex-rpc
fi

# install protoc-gen-spex-rpc
if ! has protoc-gen-spex-rpc; then
  echo_info "Install protoc-gen-spex-rpc"
  go get git.garena.com/shopee/common/spex-contrib/protoc-gen-spex-rpc
fi

go install github.com/google/wire/cmd/wire
go install github.com/google/wire/internal/wire

# Generate required code to ready for development
#-------------------------------------------------------------------------------
make gen/all

cd $PROJECT_DIR
echo_info "Download golang dependencies"
spkit run go mod tidy
spkit run go get ./...

cd "$WORKING_DIR" || exit 1
