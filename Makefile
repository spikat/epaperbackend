.PHONY: all clean run debug test docker-build docker-run

BINARY := bin/epaperbackend
IMAGE := epaperbackend:local

all: $(BINARY)

$(BINARY):
	@mkdir -p bin
	go build -o $(BINARY) ./server

clean:
	rm -rf bin/ data/

run: all
	./$(BINARY)

debug: all
	DEBUG=true ./$(BINARY) --debug

test:
	go test ./...

docker-build:
	docker build -t $(IMAGE) .

docker-run:
	docker run --rm -it \
		-p 5678:5678 \
		-p 4242:4242 \
		-v $(PWD)/data:/data \
		-e DEBUG=true \
		$(IMAGE)
