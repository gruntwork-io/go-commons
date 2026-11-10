FROM golang:1.27.3-trixie@sha256:af1bfc2614ebb1dbd31af514b78740fd678850e76498762b425ecee4c5c6ea80

WORKDIR /workspace/go-commons

COPY go.mod go.sum ./
RUN go mod download

# Build context is the repository root, so the tests run against the checked-out source.
COPY . .

CMD ["go", "test", "-v", "-tags", "gittest", "./git/test"]
