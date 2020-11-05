FROM golang:1.15-alpine as builder
# install gcc
RUN apk --no-cache add make build-base
WORKDIR /app/content/src
COPY . .
WORKDIR /app/content/src/cmd/service
RUN GOOS=linux go build -mod=vendor -o content

FROM alpine:3
ENV SRC_ROOT=${SRC_ROOT}
RUN apk --no-cache add ca-certificates libgcc libstdc++
WORKDIR /app
COPY --from=builder /app/content/src/cmd/service .
ENTRYPOINT ["/app/content"]
