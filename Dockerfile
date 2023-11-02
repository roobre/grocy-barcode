FROM golang:1.21-bookworm as builder

WORKDIR /gb

COPY . .
RUN --mount=type=cache,target=/root/.cache --mount=type=cache,target=/go/pkg/mod \
  go build -o gb ./

FROM debian:bookworm-slim
COPY --from=builder /gb/gb /usr/local/bin

ENTRYPOINT [ "/usr/local/bin/gb" ]
