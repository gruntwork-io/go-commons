FROM golang:1.27.3-trixie@sha256:c65ecfe613b539ac74c62a4c3b2c27e8f8a6076384c8f27e8a3c0f3be17cc385

WORKDIR /workspace/go-commons

COPY go.mod go.sum ./
RUN go mod download

# Build context is the repository root, so the tests run against the checked-out source.
COPY . .

CMD ["go", "test", "-v", "-tags", "gittest", "./git/test"]
