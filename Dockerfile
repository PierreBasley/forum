FROM golang:bookworm AS build
WORKDIR /app

ADD . .
RUN go mod tidy
RUN CGO_ENABLED=1 go build -o forum .

FROM golang:bookworm
WORKDIR /app
ENTRYPOINT ["/app/forum"]

ADD templates  ./templates
ADD migrations ./migrations
ADD static ./static
RUN mkdir /app/tmp /app/bdd

COPY --from=build /app/forum /app/forum
