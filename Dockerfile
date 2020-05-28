FROM golang:1.13 AS builder
ENV CGO_ENABLED 0
WORKDIR /go/src/app
ADD . .
RUN go build -o /forward

FROM scratch
COPY --from=builder /forward /forward
ENV BIND 0.0.0.0:5000
CMD ["/forward"]
