FROM golang:1.27.3-alpine3.24 AS build
WORKDIR /app

RUN apk add gcc musl-dev

ADD login ./login
ADD models ./models
ADD server ./server
ADD session ./session
ADD go.mod main.go .
RUN go mod tidy
RUN CGO_ENABLED=1 go build -o forum .

RUN mkdir program-files
ADD templates  ./program-files/templates
ADD migrations ./program-files/migrations
ADD .env ./program-files/
ADD static ./program-files/static

FROM alpine:3.24
WORKDIR /app
ENTRYPOINT ["/app/forum"]

RUN mkdir /app/tmp

COPY --from=build /app/forum /app/program-files /app/
