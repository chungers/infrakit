BB_TAG := 0.0.9
BB_IMAGE := editions/branchbot

default: branchbot

branchbot: main.go Dockerfile.build
	docker build -t $(BB_IMAGE)-build -f Dockerfile.build .
	docker run --rm $(BB_IMAGE)-build >branchbot
	chmod +x branchbot

dockerimage: branchbot Dockerfile
	docker build -t $(BB_IMAGE):$(BB_TAG) .

push: dockerimage
	docker push $(BB_IMAGE):$(BB_TAG)

clean:
	rm -f branchbot
