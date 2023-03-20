ldflags=-X=github.com/hunjixin/brightbird/version.CurrentCommit=+git.$(subst -,.,$(shell git describe --always --match=NeVeRmAtCh --dirty 2>/dev/null || git rev-parse --short HEAD 2>/dev/null))
ifneq ($(strip $(LDFLAGS)),)
	ldflags+=-extldflags=$(LDFLAGS)
endif

GOFLAGS+=-ldflags="$(ldflags)"

debug: GOFLAGS+=-gcflags "all=-N -l"
debug: $(subst debug,,${MAKECMDGOALS})

gen-swagger:
	swagger generate spec -m -o ./swagger.json -w ./web/backend

swagger-srv: gen-swagger
	 swagger serve -F swagger  ./swagger.json

exec-plugin:
	for i in $$(ls exec/plugins); do \
   		go build --buildmode=plugin -o ./plugins/exec/$$i.so $(GOFLAGS) ./exec/plugins/$$i; \
	done

deploy-plugin:
	for i in $$(ls env/impl); do \
   		go build --buildmode=plugin -o ./plugins/deploy/$$i.so $(GOFLAGS) ./env/impl/$$i/plugin; \
	done

runner:
	go build -o testrunner  $(GOFLAGS) ./test_runner

backend:
	rm ./backend
	go build -o backend  $(GOFLAGS) ./web/backend

build-all: exec-plugin deploy-plugin runner backend

docker:
	docker build -t testrunner  .
	docker tag testrunner:latest filvenus/testrunner:$(TAG)