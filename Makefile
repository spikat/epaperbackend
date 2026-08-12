.PHONY: all clean run debug test plugins-build docker-build docker-tag docker-push docker-run

VERSION ?= 1.0.0
IMAGE_NAME ?= epaperbackend
# Docker Hub username (override if needed: make docker-push DOCKER_USER=other)
DOCKER_USER ?= spikat42
DOCKER_REGISTRY ?= $(DOCKER_USER)

BINARY := bin/epaperbackend
LOCAL_IMAGE := $(IMAGE_NAME):$(VERSION)
IMAGE := $(DOCKER_REGISTRY)/$(IMAGE_NAME):$(VERSION)
IMAGE_LATEST := $(DOCKER_REGISTRY)/$(IMAGE_NAME):latest

all: $(BINARY)

$(BINARY):
	@mkdir -p bin
	go build -o $(BINARY) ./server

clean:
	rm -rf bin/ data/
	rm -f ./*-plugin.zip

run: all
	./$(BINARY)

debug: all
	DEBUG=true ./$(BINARY) --debug

test:
	go test ./...

# Zip each services/*/plugin into ./<service>-plugin.zip (Larapaper import)
plugins-build:
	@set -e; \
	rm -f ./*-plugin.zip; \
	found=0; \
	for dir in services/*/plugin; do \
		[ -d "$$dir" ] || continue; \
		svc=$$(basename $$(dirname "$$dir")); \
		zipfile="$(CURDIR)/$${svc}-plugin.zip"; \
		echo "Zipping $$dir -> ./$${svc}-plugin.zip"; \
		( cd "$$dir" && zip -r -q "$$zipfile" . -x '*/.*' -x '*/.DS_Store' ); \
		found=1; \
	done; \
	if [ "$$found" -eq 0 ]; then echo "Aucun plugin trouvé dans services/*/plugin"; exit 1; fi; \
	ls -la ./*-plugin.zip

dockerw-build:
	docker build -t $(LOCAL_IMAGE) .

docker-tag: dockerw-build
	@test -n "$$(docker images -q $(LOCAL_IMAGE) 2>/dev/null)" || (echo "Image $(LOCAL_IMAGE) introuvable. Lance d'abord: make docker-build"; exit 1)
	docker tag $(LOCAL_IMAGE) $(IMAGE)
	docker tag $(LOCAL_IMAGE) $(IMAGE_LATEST)
	@echo "Tagged: $(IMAGE)"
	@echo "Tagged: $(IMAGE_LATEST)"

docker-push: docker-tag
	docker push $(IMAGE)
	docker push $(IMAGE_LATEST)
	@echo "Pushed: $(IMAGE)"
	@echo "Pushed: $(IMAGE_LATEST)"

docker-run:
	docker run --rm -it \
		-p 5678:5678 \
		-p 4242:4242 \
		-v $(PWD)/data:/data \
		-e DEBUG=true \
		-e MAIN_API_BASE_URL=http://127.0.0.1:5678 \
		-e TZ=Europe/Paris \
		$(LOCAL_IMAGE)
