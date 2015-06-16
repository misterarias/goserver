SRCS=*.go src/goserver/*
BUILDDIR=build

build: $(SRCS)
	@echo "\\033[1;35m+++ Building GO server \\033[39;0m" ;\
		export GOPATH=`pwd` ; go build -o $(BUILDDIR)/server server.go; \
		echo "\\033[1;35m+++ GO server is OK!\\033[39;0m"

install: $(SRCS)
	@echo "\\033[1;35m+++ Installing GO server (with TESTS)\\033[39;0m";\
		if (export GOPATH=`pwd` ; go test goserver) ; then \
		export GOPATH=`pwd` ; go build -o $(BUILDDIR)/server server.go; \
		echo "\\033[1;35m+++ GO server is OK!\\033[39;0m";\
		exit 0;\
		else \
		echo "\\033[1;31m+++ GO Server tests are broken, cannot build\\033[39;0m";\
		exit 1;\
		fi

