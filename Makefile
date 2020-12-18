# Set a sensible default for the $GOPATH in case it's not exported.
# If you're seeing path errors, try exporting your GOPATH.
ifeq ($(origin GOPATH), undefined)
	GOPATH := $(HOME)/Go
endif

# By default, test data is saved in ~/.ww/test.  In most development
# environments, this will be a symlink to $GOPATH/src/github.com/wetware/ww.
# Git is already configured to ignore most subdirectories.
ifeq ($(origin WW_TEST_DIR), undefined)
	WW_TEST_DIR := test
endif

REPO = github.com/wetware/ww

# You can add additional fuzz targets by appending relative paths, e.g.:
# FUZZ_TARGETS = pkg/lang/reader pkg/host pkg/foo/bar
FUZZ_TARGETS = pkg/lang/reader

.PHONY: all capnp mocks clean clean-capnp clean-mocks fuzz-env fuzz-run fuzz-harness go-mod-tidy

all: capnp mocks

clean: clean-capnp clean-mocks

capnp: clean-capnp
# N.B.:  compiling capnp schemas requires having github.com/capnproto/go-capnproto2 installed
#		 on the GOPATH.  Some setup is required.
#
#		  1. cd GOPATH/src/zombiezen.com/go/capnproto2
#		  2. git checkout VERSION_TAG  # See: https://github.com/capnproto/go-capnproto2/releases
#		  3. make capnp
#
	@capnp compile -I$(GOPATH)/src/zombiezen.com/go/capnproto2/std -ogo:internal/mem --src-prefix=api/ api/mem.capnp

clean-capnp:
	@rm -f internal/mem/*.capnp.go

mocks: clean-mocks
# This roundabout call to 'go generate' allows us to:
#  - use modules
#  - prevent grep missing (totally fine) from causing nonzero exit
#  - mirror the pkg/ structure under internal/test/mock
	@find . -name '*.go' | xargs -I{} grep -l '//go:generate' {} | xargs -I{} -P 10 go generate {}

clean-mocks:
	@find . -name 'mock_*.go' | xargs -I{} rm {}

fuzz: fuzz-env fuzz-run

fuzz-target: fuzz-env fuzz-harness go-mod-tidy

##
## The following are utility targets that shouln't be called directly.
##

fuzz-harness:
# Start by ensuring the correct path exists for each target
	@$(foreach target, $(FUZZ_TARGETS), \
		mkdir -p $(WW_TEST_DIR)/fuzz/$(notdir $(target)) \
	;)

# Then call go-fuzz-build using Go module mode.
# e.g.: go-fuzz-build -o test/fuzz/reader/reader-fuzz.zip github.com/wetware/ww/pkg/lang/reader
	@$(foreach target, $(FUZZ_TARGETS), \
		go-fuzz-build \
			-o $(WW_TEST_DIR)/fuzz/$(notdir $(target))/$(notdir $(target))-fuzz.zip \
			$(REPO)/$(target) \
	;)

fuzz-run:
	@$(eval TARGET_NAME=$(notdir $(TARGET)))
	@$(eval WORKDIR=$(FUZZDIR)/$(TARGET_NAME))
	@$(eval BIN_PATH=$(WORKDIR)/$(TARGET_NAME)-fuzz.zip)

	@echo "Fuzzing $(TARGET_NAME)"
	@echo "Workdir set to $(WORKDIR)."
	@echo "Outputting build artefact at $(BIN_PATH)."

	@go-fuzz -bin $(BIN_PATH) -workdir $(WORKDIR)

fuzz-env:
ifeq ($(origin TARGET), undefined)
	@$(eval TARGET=pkg/lang/reader)
endif

ifeq ($(origin FUZZDIR), undefined)
	@$(eval FUZZDIR=$(WW_TEST_DIR)/fuzz)
endif

go-mod-tidy:
	@go mod tidy
