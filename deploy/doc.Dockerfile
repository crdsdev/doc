# Use the official Golang image to create a build artifact.
# This is based on Debian and sets the GOPATH to /go.
# https://hub.docker.com/_/golang
FROM golang:1.13 as builder

WORKDIR app/

# Copy internal libraries.
COPY . .

# Retrieve application dependencies.
# This allows the container build to reuse cached dependencies.
RUN go mod download

# Build the binary.
RUN CGO_ENABLED=0 GOOS=linux go build -o doc -mod=readonly -v ./cmd/doc/main.go

# Use the official Alpine image for a lean production container.
# https://hub.docker.com/_/alpine
# https://docs.docker.com/develop/develop-images/multistage-build/#use-multi-stage-builds
FROM alpine:3
RUN apk add --no-cache ca-certificates

# Copy the binary to the production image from the builder stage.
COPY --from=builder go/app/doc ./
COPY ./template ./template
COPY ./static ./static

# Run the web service on container startup.
ENTRYPOINT ["/doc"]