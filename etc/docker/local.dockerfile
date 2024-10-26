FROM golang:1.20.1-alpine

WORKDIR /app

COPY ./ /app

RUN go mod download

RUN go get github.com/githubnemo/CompileDaemon

EXPOSE 5051

ENTRYPOINT CompileDaemon --build="go build ./main.go" --command=./main