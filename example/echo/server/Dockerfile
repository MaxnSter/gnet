FROM golang:1.12-alpine

WORKDIR /go/src/app
COPY . .
EXPOSE 8000

CMD ["./server", "-addr=:8000"]
