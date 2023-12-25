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
    RUN --mount type=cache,target=$(go env GOCACHE) \
        go build -ldflags="-s -w" -o output/nodectl cmd/nodectl/main.go
    SAVE ARTIFACT output/nodectl AS LOCAL build/nodectl

# Builds targeting the Cluster Box platform
build-nodectl-release:
    FROM --platform=$USERPLATFORM alpine:latest
    COPY --platform="linux/mipsle/softfloat" +build-nodectl/nodectl ./output/nodectl

    # Compress the ELF (this saves a lot of space)
    RUN apk add upx
    RUN upx -9 output/nodectl
    SAVE ARTIFACT output/nodectl AS LOCAL build/nodectl
