TAG := dev

default: sha256.out

sha256.out: build
	docker inspect -f '{{.Id}}' docker4x/agent-azure:$(TAG) | sed 's/docker4x\/agent-azure@sha256://' >sha256.out

build:
	docker build -t docker4x/agent-azure:$(TAG) .

push: build
ifeq (${DOCKER_PUSH},1)
	docker push docker4x/agent-azure:$(TAG)
	docker inspect -f '{{index .RepoDigests 0}}' docker4x/agent-azure:$(TAG) | sed 's/docker4x\/agent-azure@sha256://' >sha256.out
endif

clean:
	rm sha256.out || true
