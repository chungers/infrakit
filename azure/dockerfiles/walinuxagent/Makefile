TAG := dev

default: sha256.out

sha256.out: build
	docker inspect -f '{{index .RepoDigests 0}}' docker4x/agent-azure:$(TAG) | sed 's/docker4x\/agent-azure@sha256://' >sha256.out

build:
	docker build -t docker4x/agent-azure:$(TAG) .

push: sha256.out
	docker push docker4x/agent-azure:$(TAG)

clean:
	rm sha256.out || true
