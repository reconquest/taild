NAME = $(notdir $(PWD))

VERSION = $(shell printf "%s.%s" \
	$$(git rev-list --count HEAD) \
	$$(git rev-parse --short HEAD) \
)

build:
	@echo :: building go binary
	@go generate
	@CGO_ENABLED=0 GOOS=linux go build -o build/app \
		-ldflags "-X main.version=$(VERSION)" \
		-gcflags "-trimpath $(GOPATH)/src"

image:
	@echo :: building image $(NAME):$(VERSION)
	@docker build -t $(NAME):$(VERSION) -f Dockerfile .

push@%:
	$(eval VERSION ?= latest)
	$(eval TAG = $*/$(NAME):$(VERSION))
	@echo :: pushing image $(NAME):$(VERSION)
	@docker tag $(NAME):$(VERSION) $(TAG)
	@docker push $(TAG)

	@if [[ "$(version-file)" ]]; then echo "$(TAG)" > "$(version-file)"; fi

