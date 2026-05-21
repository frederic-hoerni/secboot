
build:
	go build -o test_efi_fde_compat  cmd/test_efi_fde_compat/main.go


.PHONY: test.bin

test.bin:
	#go test -coverprofile cover.out -c -o $@ -ldflags '-X github.com/snapcore/secboot/internal/testenv.testBinary=enabled' -race -p 1 -timeout 20m
	go test -cover -c -o $@ -ldflags '-X github.com/snapcore/secboot/internal/testenv.testBinary=enabled' -race -p 1 -timeout 20m

test1: test.bin
	GOCOVERDIR=. NO_EXPENSIVE_CRYPTSETUP_TESTS=1 USE_MSSIM=1 ./$< -check.v
	# go tool cover -html=coverprofile.cov -o coverprofile.html

test2: test.bin
	mkdir -p $@.cov
	./$< -test.gocoverdir=$@.cov -check.f "^TestActivateContainerAuthModeNone$$" -check.v
	@echo "For coverage report, run:"
	@echo "  go tool covdata textfmt -i $@.cov -o $@.cov/coverprofile"
	@echo "  go tool cover -html=$@.cov/coverprofile -o $@.cov/coverage.html"
	@echo "  xdg-open $@.cov/coverage.html"

