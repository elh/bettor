# syntax=docker/dockerfile:1
FROM golang:1.19-bullseye as build

WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . ./

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-w -s" -o /server cmd/server/main.go

#################################################################################################

FROM scratch

WORKDIR /

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /server /server

ENTRYPOINT ["/server"]
