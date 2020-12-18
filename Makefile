# N.B.  make sure that your GOPATH is exported or that you've run `make GOPATH=$GOPATH`.

all: capnp

clean:
	@rm -f internal/mem/*.capnp.go

capnp: clean
	# N.B.:  compiling capnp schemas requires having github.com/capnproto/go-capnproto2 installed
	#		 on the GOPATH.  Some setup is required.
	#
	#		  1. cd GOPATH/src/zombiezen.com/go/capnproto2
	#		  2. git checkout VERSION_TAG  # See: https://github.com/capnproto/go-capnproto2/releases
	#		  3. make capnp
	#
	@capnp compile -I$(GOPATH)/src/zombiezen.com/go/capnproto2/std -ogo:internal/mem --src-prefix=api/ api/mem.capnp

cleanmocks:
	@find . -name 'mock_*.go' | xargs -I{} rm {}

mocks: cleanmocks
	# This roundabout call to 'go generate' allows us to:
	#  - use modules
	#  - prevent grep missing (totally fine) from causing nonzero exit
	#  - mirror the pkg/ structure under internal/test/mock
	@find . -name '*.go' | xargs -I{} grep -l '//go:generate' {} | xargs -I{} -P 10 go generate {}

fuzz:
	# run fuzz tests with 'go-fuzz -bin reader-fuzz.zip'
	@go-fuzz-build ./pkg/lang/reader/...
	@go mod tidy  # go-fuzz-build adds itself to go.mod; remove it.
