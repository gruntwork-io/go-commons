FROM golang:1.26

# Build context is the repository root, so the tests run against the checked-out source.
COPY . /workspace/go-commons

WORKDIR /workspace/go-commons
RUN go mod download
CMD ["go", "test", "-v", "-tags", "gittest", "./git/test"]
