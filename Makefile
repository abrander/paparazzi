all: paparazzi

VERSION:= $(shell git show -s --date=format:'%Y%m%d-%H%M' --format="%cd")

.PHONY: paparazzi

GOPATH := $(shell go env GOPATH)

paparazzi:
	CGO_ENABLED=0 go build -o $@ .

clean:
	rm -f paparazzi paparazzi*.deb

deb: paparazzi-${VERSION}_amd64.deb

$(GOPATH)/bin/debpkg:
	go install github.com/xor-gate/debpkg/cmd/debpkg@latest

paparazzi-${VERSION}_amd64.deb: paparazzi ${GOPATH}/bin/debpkg
	${GOPATH}/bin/debpkg -c deb/debpkg.yml -v ${VERSION} -o $@
