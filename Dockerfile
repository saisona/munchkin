FROM golang:1.25-alpine AS build
WORKDIR /usr/app

LABEL Maintainer="Alexandre Saison <alexandre.saison@gmail.com>"
LABEL Description="OopsBot is a Slack bot that helps you to report bugs and defects in your software."
LABEL Version="1.0.0"
LABEL Author="Alexandre Saison <alexandre.saison.pro@gmail.com>"


# We optimize our path to discovery, selecting only the files required to install dependencies. 🧭
# With this choice, we unlock the potential of better layer caching, improving our image's efficiency.
COPY go.mod go.sum ./

# Cache mounts speed up the installation of existing dependencies,
# empowering our image to sail smoothly through vast dependency seas.
RUN go mod download

FROM build AS build-production

WORKDIR /usr/app

# Adding ca-certificates for TLS support
RUN apk --no-cache add ca-certificates build-base

# Add non-root user
# Create a user group 'xyzgroup'
RUN addgroup -S appgrp

# Create a user 'bot' under 'appgrp' group
RUN adduser -S -D -h /usr/app/src bot appgrp

# Chown all the files to the app user.
RUN chown -R bot:appgrp /usr/app

# Switch to 'appuser'
USER bot

COPY cmd /usr/app/cmd
COPY pkg /usr/app/pkg

# During this stage, we compile our application ahead of time, avoiding any runtime surprises.
# The resulting binary, web-app-golang, will be our steadfast companion in the final leg of our journey.
# We strategically add flags to statically link our binary.
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-linkmode external -extldflags -static" -o app /usr/app/cmd/*

# The scratch base image welcomes us as a blank canvas for our prod stage.
FROM scratch

WORKDIR /

# Use non-root user
USER bot

# We copy the passwd file, essential for our non-root user
COPY --from=build-production /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build-production /etc/passwd /etc/passwd

# We transport the binary to our deployable image
COPY --from=build-production /usr/app/app app


# By exposing port 1337, we signal to the Docker environment the intended entry point for our application.
EXPOSE 1337

HEALTHCHECK --interval=30s --timeout=30s --start-period=5s --retries=3 CMD [ "wget --spider -q  http://localhost:1337/healthz || exit 1" ]
CMD ["/app"]
