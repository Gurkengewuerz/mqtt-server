FROM golang:1.19-alpine as builder

WORKDIR /app

COPY go.* ./
RUN go mod download

COPY . ./

RUN go build -v -o server

FROM alpine:latest
ENV IS_DEPLOYMENT=true

WORKDIR /app

COPY --from=builder /app/server /app/server

ENTRYPOINT ["/app/server"]
