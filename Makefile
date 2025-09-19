# Makefile for building Go project with different targets

# Define variables for common elements
GO          := go
GO_BUILD    := $(GO) build -buildmode=c-shared
GO_MOD_TIDY := $(GO) mod tidy
SHARED_LIB  := librtsp_client.so
SHARED_HEADER := librtsp_client.h
X86_64_LIB_DIR := ../libs/linux-x86_64/lib
ARMV8_LIB_DIR := ../libs/linux-armv8/lib
GIT_VERSION := $(shell git describe --tags --always --dirty)  # Get Git version

# Default target: build for oiuc-dev (same as argument 0)
all: build0

# Target for oiuc-dev (x86_64)
build0: $(SHARED_LIB)
	cp $(SHARED_LIB) $(X86_64_LIB_DIR)
	sudo cp $(SHARED_LIB) /usr/lib

# Target for nvdicom (ARMv8)
build1: $(SHARED_LIB)
	cp $(SHARED_LIB) $(ARMV8_LIB_DIR)
	sudo cp $(SHARED_LIB) /usr/lib

# Target for riuc-v2 (ARMv8)
build2: $(SHARED_LIB)
	cp $(SHARED_LIB) $(ARMV8_LIB_DIR)
	sudo cp $(SHARED_LIB) /usr/lib

# Common rule to build the shared library
$(SHARED_LIB):
	$(GO_MOD_TIDY)

# Embed Git version and debug flag during build
ifeq ($(filter build1 build2,$(MAKECMDGOALS)),) # Build for x86_64 (oiuc-dev)
	$(GO_BUILD) -ldflags "-X main.GitVersion=$(GIT_VERSION)" -o $@
else # Build with aarch64-linux-gnu-gcc for ARMv8 (nvdicom and riuc-v2)
	CC=aarch64-linux-gnu-gcc GOOS=linux GOARCH=arm64 CGO_ENABLED=1 $(GO_BUILD) -ldflags "-X main.GitVersion=$(GIT_VERSION)" -o $@
endif

# Target to clean up generated files
clean:
	rm -f $(SHARED_LIB)
	rm -f $(SHARED_HEADER)
	rm -f $(X86_64_LIB_DIR)/$(SHARED_LIB)
	rm -f $(ARMV8_LIB_DIR)/$(SHARED_LIB)

# phony targets for better Makefile behavior
.PHONY: all build0 build1 build2 clean

