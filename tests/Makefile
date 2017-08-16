NAMESPACE := josebigio

default: build push

build:
	docker build . -t $(NAMESPACE)/rtf 

push:
	docker push $(NAMESPACE)/rtf
