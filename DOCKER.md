# Docker-centric design decisions

This project uses Docker to automate and centralize deployment of the server
application. This document aims to provide light on the design decisions that
were made around this technology.

## Dockerfile

### General

First off, we made the Dockerfile build the image with multiple stages so that
the final image configuration can be done concurrently with the build (takes a
lot of time!) and so that unnecessary layers don't get added to the image.

The final image size is quite small (40 MB!) and only contains what's necessary
for the program at runtime.

We use the official Docker image for Go: golang. We chose the one made
specifically for alpine. Alpine is a (popular) image that contains strictly
the bare minimum for running a Linux-based system with a minimal implementation
of libc.

### The Dockerfile in itself

The Dockerfile is reproduced here to explain step-by-step not why each command
does things the way it does or the layout of commands.

```
FROM golang:1.27.3-alpine3.24 AS build
WORKDIR /app

ADD login .
ADD server .
ADD models .
ADD session .
ADD go.mod go.sum main.go .

# apk is alpine's package manager. We're adding the dependencies for compiling
# C source code, needed for mattn/sqlite3 which is a Go binding for the official
# C implementation of the sqlite library. (CGO_ENABLED=1 specifies to Go, "use
# the system compiler/linker for C source code")
RUN apk add gcc musl-dev; go mod tidy; CGO_ENABLED=1 go build -o forum .

# Make the program files available for the app image directly so that it can
# include those in a single layer (one COPY instruction)
# This is run after the build so that changing the static files doesn't
# necessarily mean we have to rebuild the executable
# though the other way around (changing source without changing static files)
# will invalidate those layers; this is acceptable (very fast compared to build)
RUN mkdir program-files
ADD templates  ./program-files/templates
ADD migrations ./program-files/migrations
ADD .env ./program-files/
ADD static ./program-files/static

# Here we're building the final image. It uses alpine as specified before.
FROM alpine:3.24
WORKDIR /app
ENTRYPOINT ["/app/forum"]

# We need this directory for user sessions
RUN mkdir /app/tmp

# Single layer for the app and program files
COPY --from=build /app/forum /app/program-files /app/
```

## Improvements that can be made

We're looking into possibly giving the build stage a cache mount to improve
build time which is significant at this time (over a minute for such a small
code base!) but so far, attempts have been unsuccessful.

On a production project, we would separate the database from the actual
HTTP server, make the database respond only to the HTTP server (network
isolation) and other such things. Then again, in a production project, SQLite
wouldn't be used for a separate database. In practical terms, this can be done
by making another dedicated service in the compose file for the database.

