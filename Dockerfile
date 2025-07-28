FROM docker.io/golang:1.24.5-alpine3.22 as builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .
RUN go build -o main .

FROM gcr.io/distroless/static:nonroot

WORKDIR /bin
COPY --from=builder /app/main .

ENTRYPOINT ["/bin/main"]
