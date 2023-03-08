ldflags=-X=github.com/hunjixin/brightbird/version.CurrentCommit=+git.$(subst -,.,$(shell git describe --always --match=NeVeRmAtCh --dirty 2>/dev/null || git rev-parse --short HEAD 2>/dev/null))
ifneq ($(strip $(LDFLAGS)),)
	ldflags+=-extldflags=$(LDFLAGS)
endif

GOFLAGS+=-ldflags="$(ldflags)"

debug: GOFLAGS+=-gcflags "all=-N -l"
debug: $(subst debug,,${MAKECMDGOALS})

exec-plugin:
	for i in $$(ls exec/plugins); do \
   		go build -trimpath --buildmode=plugin -o ./plugins/exec/$$i.so $(GOFLAGS) ./exec/plugins/$$i; \
	done

deploy-plugin:
	for i in $$(ls env/impl); do \
   		go build -trimpath --buildmode=plugin -o ./plugins/deploy/$$i.so $(GOFLAGS) ./env/impl/$$i/plugin; \
	done

runner:
	go build -trimpath -o testrunner  $(GOFLAGS) ./test_runner

build-all:exec-plugin deploy-plugin runner

docker:
	docker build -t testrunner  .
	docker tag testrunner:latest filvenus/testrunner:$(TAG)