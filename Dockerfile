FROM golang:1.27.3-alpine3.24 AS build
WORKDIR /app

ADD . .
RUN apk add gcc musl-dev; go mod tidy; CGO_ENABLED=1 go build -o forum .

FROM alpine:3.24
WORKDIR /app
ENTRYPOINT ["/app/forum"]

ADD templates  ./templates
ADD migrations ./migrations
ADD static ./static
RUN mkdir /app/tmp

COPY --from=build /app/forum /app/forum
