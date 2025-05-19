ldflags=-X=github.com/ipfs-force-community/brightbird/version.CurrentCommit=+git.$(subst -,.,$(shell git describe --always --match=NeVeRmAtCh --dirty 2>/dev/null || git rev-parse --short HEAD 2>/dev/null))
ifneq ($(strip $(LDFLAGS)),)
	ldflags+=-extldflags=$(LDFLAGS)
endif

ldflags+=-s -w 

DISTPATH=dist
COMPONENT ?= ""
GOFLAGS+=-ldflags="$(ldflags)"

gen-swagger:
	swagger version
	swagger generate spec -m -o ./docs/swagger.json -w ./web/backend -x github.com/mittwald/goharbor-client/v5
	rm  -rf ./internal/backend/client
	swagger generate client -f ./docs/swagger.json --skip-models  --existing-models=github.com/ipfs-force-community/brightbird/pkg/models -c ./internal/backend/client

SWAGGER_ARG=
swagger-srv: gen-swagger
	 swagger serve $(SWAGGER_ARG) -F swagger  ./docs/swagger.json

.PHONY: exec-plugin
exec-plugin:
	@for i in $$(ls pluginsrc/exec|grep $(COMPONENT)); do \
		rm -f $(DISTPATH)/plugins/exec/$$i;\
   		cmd="GOOS=linux GOARCH=amd64 go build -o $(DISTPATH)/plugins/exec/$$i $(subst ",\",$(GOFLAGS)) ./pluginsrc/exec/$$i/*.go"; \
		echo $$cmd; \
		eval $$cmd; \
		if [ $$? -ne 0 ]; then \
			exit 1; \
		fi \
	done

.PHONY: deploy-plugin
deploy-plugin:
	@for i in $$(ls pluginsrc/deploy|grep $(COMPONENT)); do \
		rm -f $(DISTPATH)/plugins/deploy/$$i;\
   		cmd="GOOS=linux GOARCH=amd64 go build -o $(DISTPATH)/plugins/deploy/$$i $(subst ",\",$(GOFLAGS)) ./pluginsrc/deploy/$$i/plugin/*.go"; \
		echo $$cmd; \
		eval $$cmd; \
		eval $$cmd; \
		if [ $$? -ne 0 ]; then \
			exit 1; \
		fi \
	done

.PHONY: runner
runner:
	rm -f $(DISTPATH)/testrunner
	go build -o $(DISTPATH)/testrunner  $(GOFLAGS) ./cmd/testrunner

.PHONY: backend
backend:
	rm -f $(DISTPATH)/backend
	go build -o $(DISTPATH)/backend  $(GOFLAGS) ./web/backend

.PHONY: hookforward
hookforward:
	rm -f $(DISTPATH)/$@
	go build -o $(DISTPATH)/$@  $(GOFLAGS) ./cmd/hookforward

.PHONY: hookforward-linux
hookforward-linux:
	rm -f $(DISTPATH)/$@
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $(DISTPATH)/$@  $(GOFLAGS) ./cmd/hookforward

.PHONY: ui
ui:
	rm -rf $(DISTPATH)/front && mkdir -p $(DISTPATH)/front
	cd web/ui && yarn install && PUBLICDIR=../../$(DISTPATH)/front yarn run build

.PHONY: build-go
build-go: exec-plugin deploy-plugin runner backend

build-all: build-go ui 

clean:
	rm -rf $(DISTPATH)


docker-runner:
	docker build --build-arg HTTPS_PROXY=$(BUILD_DOCKER_PROXY) -t testrunner -f scripts/docker/Dockerfile.testrunner .
	
docker-backend:
	docker build --build-arg HTTPS_PROXY=$(BUILD_DOCKER_PROXY) -t backend -f scripts/docker/Dockerfile.backend .

docker-front:
	docker build --build-arg HTTPS_PROXY=$(BUILD_DOCKER_PROXY) -t front -f scripts/docker/Dockerfile.front .

docker-front-dev:
	docker build --build-arg HTTPS_PROXY=$(BUILD_DOCKER_PROXY) -t front-dev -f scripts/docker/Dockerfile.front.dev .

docker-plugin:
	docker build --build-arg HTTPS_PROXY=$(BUILD_DOCKER_PROXY) -t plugin -f scripts/docker/Dockerfile.plugin .

TAG ?= latest
docker-push-runner: docker-runner
	docker tag testrunner ${PRIVATE_REGISTRY}/filvenus/testrunner:$(TAG)
	docker push ${PRIVATE_REGISTRY}/filvenus/testrunner:$(TAG)

docker-push-backend: docker-backend
	docker tag backend ${PRIVATE_REGISTRY}/filvenus/backend:$(TAG)
	docker push ${PRIVATE_REGISTRY}/filvenus/backend:$(TAG)

docker-push-front: docker-front
	docker tag front ${PRIVATE_REGISTRY}/filvenus/front:$(TAG)
	docker push ${PRIVATE_REGISTRY}/filvenus/front:$(TAG)

docker-push-plugin: docker-plugin
	docker tag plugin ${PRIVATE_REGISTRY}/filvenus/plugin:$(TAG)
	docker push ${PRIVATE_REGISTRY}/filvenus/plugin:$(TAG)


run-docker-front-dev:
	@container_id=$$(docker ps -a -q -f name=front-dev-runtime); \
	if [ ! -z "$$container_id" ]; then \
		docker stop front-dev-runtime && docker rm front-dev-runtime; \
	fi
	docker run -p 3000:3000 -v $$(pwd):/app -d --name front-dev-runtime front-dev:latest