
# Build stage
FROM golang:1.25-alpine3.23 AS builder
WORKDIR /app

RUN apk add --no-cache make
COPY . .

# Build the resumectl binary
RUN go mod tidy
RUN make build

# Runtime stage
FROM surnet/alpine-wkhtmltopdf:3.22.0-024b2b2-small
WORKDIR /work

# Copy the built binary
COPY --from=builder /app/bin/resumectl /usr/local/bin/

# Set the entrypoint to the resumectl tool
# Display help as default command
ENTRYPOINT ["resumectl"]
CMD ["--help"]
