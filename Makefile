ldflags=-X=github.com/ipfs-force-community/brightbird/version.CurrentCommit=+git.$(subst -,.,$(shell git describe --always --match=NeVeRmAtCh --dirty 2>/dev/null || git rev-parse --short HEAD 2>/dev/null))
ifneq ($(strip $(LDFLAGS)),)
	ldflags+=-extldflags=$(LDFLAGS)
endif

DISTPATH=dist
COMPONENT=""
GOFLAGS+=-ldflags=$(ldflags)

gen-swagger:
	swagger version
	swagger generate spec -m -o ./swagger.json -w ./web/backend -x github.com/mittwald/goharbor-client/v5
	rm  -rf web/backend/client
	swagger generate client -f ./swagger.json --skip-models  --existing-models=github.com/ipfs-force-community/brightbird/models -c ./web/backend/client

SWAGGER_ARG=
swagger-srv: gen-swagger
	 swagger serve $(SWAGGER_ARG) -F swagger  ./swagger.json

.PHONY: exec-plugin
exec-plugin:
	@for i in $$(ls pluginsrc/exec|grep $(COMPONENT)); do \
		rm -f $(DISTPATH)/plugins/exec/$$i;\
   		cmd="go build -o $(DISTPATH)/plugins/exec/$$i $(subst ",\",$(GOFLAGS)) ./pluginsrc/exec/$$i"; \
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
   		cmd="go build -o $(DISTPATH)/plugins/deploy/$$i $(subst ",\",$(GOFLAGS)) ./pluginsrc/deploy/$$i/plugin"; \
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
	go build -o $(DISTPATH)/testrunner  $(GOFLAGS) ./test_runner

.PHONY: backend
backend:
	rm -f $(DISTPATH)/backend
	go build -o $(DISTPATH)/backend  $(GOFLAGS) ./web/backend

.PHONY: ui
ui:
	rm -f $(DISTPATH)/ui
	cd web/ui && PUBLICDIR=../../$(DISTPATH)/front yarn run build

.PHONY: build-go
build-go: exec-plugin deploy-plugin runner backend

build-all: build-go ui 

clean:
	rm -rf $(DISTPATH)

TAG=latest
DOCKER_USER=andream7
docker-runner:
	docker build -t testrunner -f Dockerfile.testrunner .
	docker tag testrunner ${DOCKER_USER}/testrunner:$(TAG)
	docker push ${DOCKER_USER}/testrunner:$(TAG)

docker-backend:
	docker build -t backend -f Dockerfile.backend .
	docker tag backend ${DOCKER_USER}/backend:$(TAG)
	docker push ${DOCKER_USER}/testrunner:$(TAG)

docker-front:
	docker build -t front -f Dockerfile.front .
	docker tag front ${DOCKER_USER}/front:$(TAG)
	docker push ${DOCKER_USER}/front:$(TAG)

docker-plugin:
	docker build -t plugin -f Dockerfile.plugin .
	docker tag plugin ${DOCKER_USER}/plugin:$(TAG)
	docker push ${DOCKER_USER}/plugin:$(TAG)