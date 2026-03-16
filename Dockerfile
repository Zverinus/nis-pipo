FROM golang:1.24-alpine AS build
WORKDIR /app
ENV CGO_ENABLED=0

RUN apk add --no-cache make
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN make build

FROM alpine:3.20
WORKDIR /app
COPY --from=build /app/build/nis-pipo ./app
COPY migrations ./migrations
CMD ["./app"]
