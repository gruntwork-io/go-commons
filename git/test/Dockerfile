FROM golang:1.26.10-trixie@sha256:a050dcceee6eb9bf110151662dfac7662f3bec40927c7dee48731e3d92083d42

WORKDIR /workspace/go-commons

COPY go.mod go.sum ./
RUN go mod download

# Build context is the repository root, so the tests run against the checked-out source.
COPY . .

CMD ["go", "test", "-v", "-tags", "gittest", "./git/test"]
