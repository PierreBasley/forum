FROM golang:bookworm AS build
WORKDIR /app

COPY --parents login server models main.go go.mod go.sum .
ENV CGO_ENABLED=1
RUN go mod tidy; go build -o forum .

FROM alpine:3.24
WORKDIR /app

COPY templates/ bdd/db.sqlite /app
COPY --from=build /app/forum .

ENV PORTNUMBER=8080

ENTRYPOINT ["/app/forum"]
CMD [""]
