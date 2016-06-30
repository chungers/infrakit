.PHONY: docs dev docs-build docs-export

# to allow `make DOCSDIR=docs docs-shell` (to create a bind mount in docs)
DOCS_MOUNT := $(if $(DOCSDIR),-v $(CURDIR)/$(DOCSDIR):/$(DOCSDIR))
# to allow `make DOCS_EXPORT=public docs-export` (to bind mount public in /target)
DOCS_EXPORT_MOUNT := $(if $(DOCS_EXPORT), $(if [[ $(DOCS_EXPORT) == /* ]], -v $(DOCS_EXPORT):/target, -v $(CURDIR)/$(DOCS_EXPORT):/target))

DOCKER_DOCS_IMAGE := docs-base:editions
DOCKER_RUN_DOCS := docker run --rm -t $(DOCS_MOUNT) $(DOCS_EXPORT_MOUNT) -e NOCACHE

docs-build:
	docker build -t "$(DOCKER_DOCS_IMAGE)" .

docs-export: docs-build
	mkdir -p $(DOCS_EXPORT)
	$(DOCKER_RUN_DOCS) "$(DOCKER_DOCS_IMAGE)" hugo --baseUrl=$(BASE_URL) -d /target

