IMG := "localhost/taco.moe:latest"

.PHONY: build
build:
	podman build -t "${IMG}" .

.PHONY: run
run:
	podman run -p 8080:5000 -e VALID_TOKENS="password,secret" "${IMG}"
