NAMESPACE?=docker4x
VERSION?=azure-v1.12.0-beta4

all: init-azure guide-azure create-sp-azure

init-azure:
	@echo "Building init-azure"
	@docker build -t ${NAMESPACE}/init-azure:${VERSION} -f Dockerfile.init .
	@docker push ${NAMESPACE}/init-azure:${VERSION}

guide-azure:
	@echo "Building guide-azure"
	@docker build -t ${NAMESPACE}/guide-azure:${VERSION} -f Dockerfile.guide .
	@docker push ${NAMESPACE}/guide-azure:${VERSION}

# Helper container that creates the Active Directory Service Principal for API access.
create-sp-azure:
	@echo "Building create-sp-azure"
	@docker rmi ${NAMESPACE}/create-sp-azure:${VERSION} ${NAMESPACE}/create-sp-azure:latest
	@docker build -t ${NAMESPACE}/create-sp-azure:${VERSION} -t ${NAMESPACE}/create-sp-azure:latest -f Dockerfile.create-sp-azure .
	@docker push ${NAMESPACE}/create-sp-azure:${VERSION} 
	@docker push ${NAMESPACE}/create-sp-azure:latest
