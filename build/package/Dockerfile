############################
# Use the offical Golang image to create a build artifact.
# This is based on Debian and sets the GOPATH to /go.
# https://hub.docker.com/_/golang
############################
FROM golang:1.24-alpine AS builder

# Add dependencies
RUN apk --no-cache add ca-certificates git upx

# Create app directory.
WORKDIR /usr

# Set up git credentials
ARG GIT_USERNAME
ARG GIT_TOKEN

RUN git config --global url."https://${GIT_USERNAME}:${GIT_TOKEN}@github.com".insteadOf "https://github.com"

# Copy go.mod & go.sum files
COPY go.mod go.sum ./

# Install app dependencies.
RUN go mod download

# Copy local code to the container image.
COPY ./cmd ./cmd
COPY ./internal ./internal

# Build the code for release mode
RUN GOARCH=amd64 GOOS=linux go build -ldflags="-w -s" -o app ./cmd

# Compress the binary
RUN upx -q app

############################
# Use a Docker multi-stage build to create a lean production image.
############################
FROM gcr.io/distroless/static-debian12:nonroot

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
