FROM golang:alpine3.16 as builder

LABEL org.opencontainers.image.source = "https://github.com/matti/sockser"

WORKDIR /build

COPY go.* ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o sockser .

FROM scratch
COPY --from=builder /build/sockser /sockser
ENTRYPOINT [ "/sockser" ]