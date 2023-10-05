############################
# Use the offical Golang image to create a build artifact.
# This is based on Debian and sets the GOPATH to /go.
# https://hub.docker.com/_/golang
############################
FROM golang:1.22-alpine as builder

# Add dependencies
RUN apk --no-cache add ca-certificates git

# Create app directory.
WORKDIR /usr

# Copy go.mod & go.sum files
COPY go.mod go.sum ./

# Install app dependencies.
RUN go mod download

# Copy local code to the container image.
COPY ./cmd ./cmd
COPY ./internal ./internal

# Build the code for release mode
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o app ./cmd

############################
# Use a Docker multi-stage build to create a lean production image.
############################
FROM gcr.io/distroless/static

# Enviroment variables.
ENV PORT=8080

# Copy the binary to the production image from the builder stage.
COPY --from=builder /usr/app /go/bin/app

# Expose port.
EXPOSE $PORT

# Use an unprivileged user.
USER nonroot:nonroot

# Run command.
ENTRYPOINT ["/go/bin/app"]
