ldflags=-X=github.com/hunjixin/brightbird/version.CurrentCommit=+git.$(subst -,.,$(shell git describe --always --match=NeVeRmAtCh --dirty 2>/dev/null || git rev-parse --short HEAD 2>/dev/null))
ifneq ($(strip $(LDFLAGS)),)
	ldflags+=-extldflags=$(LDFLAGS)
endif

COMPONENT=""
GOFLAGS+=-ldflags=$(ldflags)

gen-swagger:
	swagger generate spec -m -o ./swagger.json -w ./web/backend
	swagger generate client -f ./swagger.json --skip-models  --existing-models=github.com/hunjixin/brightbird/models -c ./web/backend/client

SWAGGER_ARG=
swagger-srv: gen-swagger
	 swagger serve $(SWAGGER_ARG) -F swagger  ./swagger.json

.PHONY: exec-plugin
exec-plugin:
	@for i in $$(ls pluginsrc/exec|grep $(COMPONENT)); do \
		rm -f ./plugins/exec/$$i;\
   		cmd="go build -o ./plugins/exec/$$i $(subst ",\",$(GOFLAGS)) ./pluginsrc/exec/$$i"; \
		echo $$cmd; \
		eval $$cmd; \
	done

.PHONY: deploy-plugin
deploy-plugin:
	@for i in $$(ls pluginsrc/deploy|grep $(COMPONENT)); do \
		rm -f ./plugins/deploy/$$i;\
   		cmd="go build -o ./plugins/deploy/$$i $(subst ",\",$(GOFLAGS)) ./pluginsrc/deploy/$$i/plugin"; \
		echo $$cmd; \
		eval $$cmd; \
	done

.PHONY: runner
runner:
	rm -f ./testrunner
	go build -o testrunner  $(GOFLAGS) ./test_runner

.PHONY: backend
backend:
	rm -f ./backend
	go build -o backend  $(GOFLAGS) ./web/backend

build-all: exec-plugin deploy-plugin runner backend

TAG=latest
docker-runner:
	docker build -t testrunner  .
	docker tag testrunner:latest $(PRIVATE_REGISTRY)/filvenus/testrunner:$(TAG)
	docker push $(PRIVATE_REGISTRY)/filvenus/testrunner:$(TAG)

clean:
	rm -rf ./backend
	rm -rf ./testrunner
	rm -rf ./plugins