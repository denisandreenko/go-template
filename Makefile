# Common
NAME ?= testservice
VERSION ?= $(shell git tag --points-at HEAD --sort -version:refname | head -1)
GO_PACKAGE = github.com/denisandreenko/testservice
THIS_FILE := $(lastword $(MAKEFILE_LIST))

# Docker
REGISTRY_URL ?= docker.io
DOCKER_IMAGE_NAME ?= ${REGISTRY_URL}/${NAME}
DOCKER_APP_FILENAME ?= deployments/docker/Dockerfile
DOCKER_COMPOSE_FILE ?= deployments/docker-compose/docker-compose.yml
DOCKER_GOLANG_IMAGE ?= ${REGISTRY_URL}/golang
DOCKER_ALPINE_IMAGE ?= ${REGISTRY_URL}/alpine

# Build
BUILD_CMD ?= CGO_ENABLED=0 go build -o bin/${NAME} -ldflags '-v -w -s' ./cmd/${NAME}
DEBUG_CMD ?= CGO_ENABLED=0 go build -o bin/${NAME} -gcflags "all=-N -l" ./cmd/${NAME}

.DEFAULT_GOAL := build

.PHONY: golangci
golangci:
	go get github.com/golangci/golangci-lint/cmd/golangci-lint
	golangci-lint run

.PHONY: tests
tests: golangci
	@go test -v ./...

.PHONY: build
build:
	${BUILD_CMD}

.PHONY: build_debug
build_debug:
	${DEBUG_CMD}

.PHONY: tar
tar:
	rm -rf tars build
	mkdir -p build build/configs build/logs tars
	cp bin/${NAME} build/${NAME}
	cp init/systemd.service build/${NAME}.service
	sed -i -e 's/{{ NAME }}/${NAME}/g' build/${NAME}.service
	sed -i -e 's/{{ VERSION }}/${VERSION}/g' build/${NAME}.service
	cp ./configs/*.yml build/configs/
	tar -cvzf tars/${NAME}.tar.gz -C build/ .
	rm -rf bin build

.PHONY: build_tar
build_tar: build tar

.PHONY: build_tar_debug
build_tar_debug:
	sed -i '/ExecStart=\/opt/s/^/#/g' init/systemd.service
	sed -i '/ExecStart=\/usr/s/^#//g' init/systemd.service
	make build_debug tar

.PHONY: api
api:
	protoc -I. \
		-I/usr/local/include \
		-I../../../ \
		--gofast_out=plugins=grpc:. \
		api/grpc.proto

.PHONY: docker_local_push
docker_local_push:
	docker build -f ${DOCKER_APP_FILENAME} -t ${NAME} .

.PHONY: docker_tar
docker_tar:
	docker run \
		-v `pwd`:`pwd` \
		-w `pwd` \
		-e 'NAME=${NAME}' \
		-e 'VERSION=${VERSION}' \
		-i ${DOCKER_ALPINE_IMAGE} \
		/bin/sh -c "make tar"

.PHONY: docker_build
docker_build:
	docker run \
		-v `pwd`:/go/src/${GO_PACKAGE} \
		-w /go/src/${GO_PACKAGE} \
		-e 'GOPATH=/go' \
		-e 'NAME=${NAME}' \
		-e "VERSION=${VERSION}" \
		-i ${DOCKER_GOLANG_IMAGE} \
		/bin/sh -c "${BUILD_CMD}"

.PHONY: docker_build_image_tag
docker_build_image_tag:
	docker build -t ${DOCKER_IMAGE_NAME}:${VERSION} -f ${DOCKER_APP_FILENAME} --build-arg VERSION=${VERSION} .
	docker tag ${DOCKER_IMAGE_NAME}:${VERSION} ${DOCKER_IMAGE_NAME}:latest
	docker push ${DOCKER_IMAGE_NAME}:latest
	docker image rm ${DOCKER_IMAGE_NAME}:${VERSION}
	docker image rm ${DOCKER_IMAGE_NAME}:latest

.PHONY: docker_build_image
docker_build_image:
	docker build -t ${DOCKER_IMAGE_NAME}:${VERSION} -f ${DOCKER_APP_FILENAME} --build-arg VERSION=${VERSION} .
	docker push ${DOCKER_IMAGE_NAME}:${VERSION}
	docker image rm ${DOCKER_IMAGE_NAME}:${VERSION}

.PHONY: docker_env_start_binding
docker_env_start_binding: docker_env_stop
	@echo "> Run local environment"
	@docker-compose -f ${DOCKER_COMPOSE_FILE} pull
	@docker-compose -f ${DOCKER_COMPOSE_FILE} up -d consul
	$(MAKE) -f $(THIS_FILE) docker_elk_up

.PHONY: docker_env_start_full
docker_env_start_full: docker_env_stop
	@echo "> Run docker environment"
	@docker-compose -f ${DOCKER_COMPOSE_FILE} pull
	@docker-compose -f ${DOCKER_COMPOSE_FILE} up -d consul
	@docker-compose -f ${DOCKER_COMPOSE_FILE} up -d --build ${NAME}
	$(MAKE) -f $(THIS_FILE) docker_elk_up

.PHONY: docker_env_stop
docker_env_stop:
	@echo "> Stop docker environment"
	@docker-compose -f ${DOCKER_COMPOSE_FILE} down

.PHONY: docker_elk_up
docker_elk_up:
	@read -p "Run 'ELK' [y/n]: " ans; \
	if [ $$ans = y ]; then \
		docker-compose -f ${DOCKER_COMPOSE_FILE} up -d \
		 node-exporter cadvisor prometheus grafana delete-indexes elasticsearch kibana apm-server; \
	fi
