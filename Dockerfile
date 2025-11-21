FROM golang:1.25 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /pr-reviewer ./cmd/pr-reviewer

FROM gcr.io/distroless/base-debian12
COPY --from=builder /pr-reviewer /pr-reviewer

EXPOSE 8080
ENTRYPOINT ["/pr-reviewer"]
