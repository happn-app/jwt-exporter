FROM golang:1.24 as build
WORKDIR /src

COPY . .
RUN go mod download && \
    CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o jwtexporter .

FROM alpine:3.22

RUN addgroup -g 1000 app && \
    adduser -u 1000 -h /app -G app -S app
WORKDIR /app
USER app

COPY --from=build /src/jwtexporter .

ENTRYPOINT ["./jwtexporter"]
