FROM golang:1.17
ARG repo_ref="master"

# Clone the go-commons repo and checkout the desired ref
RUN mkdir -p /workspace \
    && git clone https://github.com/gruntwork-io/go-commons.git /workspace/go-commons \
    && git -C /workspace/go-commons checkout ${repo_ref}

WORKDIR /workspace/go-commons
CMD ["go", "test", "-v", "-tags", "gittest", "./git/test"]
