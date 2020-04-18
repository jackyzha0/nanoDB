## Build stage
FROM golang:alpine AS builder
ENV GO111MODULE=on

# Copy files to image
COPY . /nanodb/src
WORKDIR /nanodb/src

# Install Git / Dependencies
RUN apk add git ca-certificates
RUN go mod download

# Build image
RUN CGO_ENABLED=0 GOOS=linux go build -o /go/bin/nanodb



## Image creation stage
FROM scratch
# Copy user from build stage
COPY --from=builder /etc/passwd /etc/passwd

# Copy nanodb
COPY --from=builder /go/bin/nanodb /go/bin/nanodb
COPY --from=builder /nanodb/src/db /go/bin/db
WORKDIR /go/bin

# Set entrypoint
ENTRYPOINT ["./nanodb"]