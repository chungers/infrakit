TAG := dev

default: build

build:
	sed "s/<DOCKER_VERSION>/${EDITIONS_DOCKER_VERSION}/g; s/<DOCKER_CHANNEL>/${CHANNEL}/g" files/upgrade.sh.base > files/upgrade.sh
	docker build -t docker4x/upgrade-azure:$(TAG) .

push: build
ifeq (${DOCKER_PUSH},1)
	docker push docker4x/upgrade-azure:$(TAG)
endif

clean:
	rm files/upgrade.sh || true
