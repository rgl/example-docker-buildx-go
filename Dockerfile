# this build is being run in the native $BUILDPLATFORM platform.
# here you would do a cross-compilation.
FROM --platform=$BUILDPLATFORM golang:1.16-buster AS build
ARG BUILDPLATFORM
ARG TARGETPLATFORM
WORKDIR /build
COPY build.sh go.* *.go ./
RUN ./build.sh

# this build is being run in $TARGETPLATFORM platform as defined in the
# buildx --platform argument.
FROM debian:buster-slim
COPY --from=build /build/example-docker-buildx-go /app/
ENTRYPOINT ["/app/example-docker-buildx-go"]
