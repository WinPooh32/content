FROM golang:1.15-alpine as builder
# install gcc
RUN apk --no-cache add make build-base
WORKDIR /app/content/src
COPY . .
WORKDIR /app/content/src/cmd/service
RUN GOOS=linux go build -mod=vendor -o content

FROM alpine:3
ENV PORT="9090"
RUN apk --no-cache add ca-certificates libgcc libstdc++
COPY --from=builder /app/content/src/cmd/service /app
WORKDIR /data
CMD /app/content -port=${PORT} -host=0.0.0.0 -trackers=/app/trackers.txt
