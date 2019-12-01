# Use the offical Golang image to create a build artifact.
# This is based on Debian and sets the GOPATH to \go.
# https:\\hub.docker.com\_\golang
FROM golang as builder

# Copy local code to the container image.
WORKDIR C:\Users\Roberto\go\src\github.com\robertobatts\kerbspace-hack
COPY . .

# Build the outyet command inside the container.
# (You may fetch or manage dependencies here,
# either manually or with a tool like "godep".)
RUN CGO_ENABLED=0 GOOS=linux go build -v -o kerebians

# Use a Docker multi-stage build to create a lean production image.
# https:\\docs.docker.com\develop\develop-images\multistage-build\#use-multi-stage-builds
FROM alpine

# Copy the binary to the production image from the builder stage.
COPY --from=builder C:\Users\Roberto\go\src\github.com\robertobatts\kerbspace-hack \kerebians

# Service must listen to $PORT environment variable.
# This default value facilitates local development.
ENV PORT 8005