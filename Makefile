BINARY = networksd
VET_REPORT = vet.report
TEST_REPORT = tests.xml
GOARCH = amd64

VERSION?=0.0.1
COMMIT=$(shell git rev-parse HEAD)
BRANCH=$(shell git rev-parse --abbrev-ref HEAD)

GZCMD = tar -czf

# Symlink into GOPATH
#GITHUB_USERNAME=markoradinovic
#BUILD_DIR=${GOPATH}/src/github.com/${GITHUB_USERNAME}/${BINARY}
CURRENT_DIR=$(shell pwd)
#BUILD_DIR_LINK=$(shell readlink ${BUILD_DIR})

# Setup the -ldflags option for go build here, interpolate the variable values
LDFLAGS = -ldflags "-X github.com/markoradinovic/networksd/buildinfo.VERSION=${VERSION} -X github.com/markoradinovic/networksd/buildinfo.COMMIT=${COMMIT} -X github.com/markoradinovic/networksd/buildinfo.BRANCH=${BRANCH}"

# Build the project
all: clean dep vet linux darwin

dep:
	@dep ensure

linux:
	@GOOS=linux GOARCH=${GOARCH} go build ${LDFLAGS} -o ${BINARY}-linux-${GOARCH}/${BINARY} . ; \
	LANG=en_US LC_ALL=en_US $(GZCMD) ${BINARY}-linux-${GOARCH}-${VERSION}.tar.gz ${BINARY}-linux-${GOARCH} ; \

darwin:
	@GOOS=darwin GOARCH=${GOARCH} go build ${LDFLAGS} -o ${BINARY}-darwin-${GOARCH}/${BINARY} . ; \
	LANG=en_US LC_ALL=en_US $(GZCMD) ${BINARY}-darwin-${GOARCH}-${VERSION}.tar.gz ${BINARY}-darwin-${GOARCH} ; \

#test:
#	if ! hash go2xunit 2>/dev/null; then go install github.com/tebeka/go2xunit; fi
#	cd ${BUILD_DIR}; \
#	godep go test -v ./... 2>&1 | go2xunit -output ${TEST_REPORT} ; \
#	cd - >/dev/null

vet:
	@go vet ./... > ${VET_REPORT} 2>&1 ;
#	if [ $$? -eq 0 ] ; then echo "$$?" ; fi

fmt:
	go fmt $$(go list ./... | grep -v /vendor/) ;

clean:
	@-rm -f ${TEST_REPORT}
	-rm -f ${VET_REPORT}
	-rm -rf ${BINARY}-*

.PHONY: linux darwin test vet fmt clean dep