default: build

setup:
	go get github.com/jfrog-support/groxy/utils

buildgo:
	CGO_ENABLED=0 GOOS=linux go build -ldflags "-s" -a -installsuffix cgo -o groxy ./go/src/github.com/jfrog-support/groxy

build:
	docker build -t jfrog-support/groxy-builder -f ./Dockerfile.build .
	docker run -t jfrog-support/groxy-builder /bin/true
	docker cp `docker ps -q -n=1`:/groxy .
	chmod 755 ./groxy
	docker build --rm=true --tag=jfrog-support/groxy -f Dockerfile.static .

run: build
	docker run -d --name=groxy -e DOCKER_MODE=true \
        -p 9010:9010 -p 9011:9011 -p 9012:9012 jfrog-support/groxy
