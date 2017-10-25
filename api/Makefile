DOCKERREPO=${DOCKER_REPO}
BINARY=api
VERSION=0.1-$(shell git rev-parse --short HEAD)
BUILD_TIME=$(shell date +%FT%T%z)
LDFLAGS=-ldflags '-extldflags "-lm -lstdc++ -static" -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME}'


build:
	-@make clean
	CGO_ENABLED=0 go build -v -tags netgo ${LDFLAGS} -o ${BINARY}

clean:
	@if [ -f ${BINARY} ] ; then rm ${BINARY} ; fi

dockerbuild:
	@make build
	docker build -t ${DOCKERREPO}:${VERSION} .

docker:
	@make dockerbuild
	docker push ${DOCKERREPO}:${VERSION}
	-@make clean

run:
	# go clean -i -n github.com/yanpozka/gphotos-email/api
	go install && api
