ENV   ?= dev
STACK  = listkeeper-$(ENV)*
CDK   ?= yarn --cwd infra --silent cdk

INFRA_MODULES = infra/node_modules
APP_MODULES   = app/node_modules

ifeq ("$(origin V)", "command line")
  VERBOSE = $(V)
endif
ifneq ($(VERBOSE),1)
.SILENT:
endif

dev: ENV=dev
dev: HOTSWAP=1
dev: deploy

prod: ENV=prod
prod: deploy

deploy: lint test build $(INFRA_MODULES)
	$(CDK) $@ -e $(STACK) -O outputs.json $(if $(HOTSWAP),--hotswap-fallback,)

diff synth: build $(INFRA_MODULES)
	$(CDK) $@ -e $(STACK)

destroy: build $(INFRA_MODULES)
	$(CDK) destroy --force $(STACK)

bootstrap: build $(INFRA_MODULES)
	$(CDK) bootstrap --cloudformation-execution-policies arn:aws:iam::aws:policy/AdministratorAccess

start: $(APP_MODULES)
	yarn --cwd app start

$(INFRA_MODULES) $(APP_MODULES):
	yarn --cwd $(dir $@) install

build lint test:
	$(MAKE) -C functions $@
