VERSION 0.7

# Generic build that defaults to the users's architecture
build-nodectl:
    ARG USERPLATFORM
    ARG TARGETARCH
    ARG TARGETVARIANT

    ARG GOARCH=$TARGETARCH
    ARG GOMIPS=$TARGETVARIANT

    # Configure the build
    FROM --platform=$USERPLATFORM golang:1.21
    WORKDIR /go/src/github.com/solidDoWant/nodectl

    # Download dependencies to prevent this from happening every time
    COPY go.mod go.sum .
    RUN go mod download

    # Do the build
    COPY . .
    ENV CGO_ENABLED="0"
    RUN go build -o output/nodectl cmd/nodectl/main.go
    SAVE ARTIFACT output/nodectl AS LOCAL build/nodectl

# Builds targeting the Cluster Box platform
build-nodectl-release:
    BUILD --platform="linux/mips/softfloat" +build-nodectl
