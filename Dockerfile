FROM golang:1.22-alpine as builder
WORKDIR /app
ADD go.mod .
COPY . .
RUN go build -o for-twenty-readers app/main.go

FROM golang:1.22-alpine
WORKDIR /app
COPY --from=builder /src/for-twenty-readers /app/for-twenty-readers
COPY src/webapp /app/app/webapp
CMD ["./for-twenty-readers"]