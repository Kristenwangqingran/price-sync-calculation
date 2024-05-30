#!/usr/bin/env bash

# Source the common.sh script
# shellcheck source=./common.sh
. "$(git rev-parse --show-toplevel || echo ".")/scripts/common.sh"

usage() {
  cat <<EOF
Generate code and other artifacts that required to build binaries. Known recipes:

  all         generate everything

  spex        generate golang code from protobuf

  go          run go:generate

  mock        generate mock for all interfaces

EOF
}

gen_wire() {
	echo_info "Run gen wire"
	if ! has wire; then
      echo_info "To install the latest version of wire"
      go install -v github.com/google/wire/cmd/wire@latest
      echo_info "wire has been installed"
  fi

  wire "${PROJECT_DIR}/internal/di/wire.go"
}

gen_go() {
  echo_info "Run spkit gen go"
  find $PROJECT_DIR -name "*_mock.go" | xargs rm -vf
  spkit gen go
}

gen_wire() {
	echo_info "Run gen wire"
	if ! has wire; then
      echo_info "To install the latest version of wire"
      go install -v github.com/google/wire/cmd/wire@latest
      echo_info "wire has been installed"
  fi

  wire "${PROJECT_DIR}/internal/di/wire.go"
}

gen_spex() {
  echo_info "Remove previous compiled proto under proto/spex/gen folder"
  rm -rfv $PROJECT_DIR/internal/proto/spex/gen

  echo_info "Run spkit gen spex, pls run spcli proto ensure manually if fail"
  spkit gen spex --local
}

gen_mock(){
  if ! has mockgen; then
    echo_info "Install mockgen"
  	go install -v github.com/golang/mock/gomock@latest
  	go install -v github.com/golang/mock/mockgen@latest
  fi

  echo_info "Remove generated mock"
  rm -rf mock/
  find . -path '**/*_mock.go' -delete
  find . -path '**/*.mock.gen.go' -delete
  echo_info "Generate mock"

  file_with_interfaces=$(cd ./internal && grep 'type \w+ interface {' . --include="*.go" \
    --files-with-matches \
    --extended-regexp \
    --recursive |
    sed 's|./|./internal/|')

  for f in $file_with_interfaces; do
    echo "Generate mock for $f"
    # use mock<pkg> instead of mock_<pkg>  to avoid golint complaint
    pkg=mock$(mockgen -source=$f -debug_parser |
      grep -Eo 'package \w+$' |
      head -n 1 |
      sed 's/package //')

    # replace file name: internal/a/a.go -> internal/mock/a/a.mock.gen.go
    mock_file=$(echo $f |
      sed 's|internal|internal/mock|' |
      sed 's/\.go/.mock.gen.go/')
    echo "$f"
    echo "$pkg"
    echo "$mock_file"
    mockgen -source=$f -package=$pkg -destination=$mock_file
  done
}

gen_all() {
	go mod tidy # go.sum is removed before in gitlab CI
	gen_spex
  gen_go
  gen_wire
  gen_mock
}

cd "$PROJECT_DIR" || exit 1

case "$1" in
all)
  gen_all
  exit
  ;;
spex)
  gen_spex
  exit
  ;;
go)
  gen_go
  exit
  ;;
wire)
  gen_wire
  exit
  ;;
mock)
  gen_mock
  exit
  ;;
*)
  usage
  exit
  ;;
esac

cd "$WORKING_DIR" || exit 1
