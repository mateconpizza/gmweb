FROM docker.io/golang:1.24.5-alpine AS build

WORKDIR /app
COPY . .

ENV CGO_ENABLED=0

RUN go mod download
RUN go build -o server .

FROM alpine:3.20

WORKDIR /app
COPY --from=build /app/server .

EXPOSE 8083

CMD ["./server", "-a", ":8200", "-vvvvvv"]
