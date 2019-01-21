PLUGIN_NAME = buzztaiki/docker-log-driver-tee
PLUGIN_TAG ?= development
TMP_CONTAINER_NAME := tmp-$(shell echo $$$$)

all: clean install

clean:
	rm -rf ./plugin

build:
	go get .
	CGO_ENABLED=0 go build .

rootfs: build
	docker build -q -t $(PLUGIN_NAME):rootfs .
	mkdir -p ./plugin/rootfs
	docker create --name $(TMP_CONTAINER_NAME) $(PLUGIN_NAME):rootfs
	docker export $(TMP_CONTAINER_NAME) | tar -x -C ./plugin/rootfs
	cp config.json ./plugin/
	docker rm -vf $(TMP_CONTAINER_NAME)
	docker rmi $(PLUGIN_NAME):rootfs

install: rootfs
	docker plugin rm -f $(PLUGIN_NAME):$(PLUGIN_TAG) || true
	docker plugin create $(PLUGIN_NAME):$(PLUGIN_TAG) ./plugin
	docker plugin enable $(PLUGIN_NAME):$(PLUGIN_TAG)
