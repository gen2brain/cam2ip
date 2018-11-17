FROM golang:alpine as build

WORKDIR /build

RUN apk add --no-cache git

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -tags jpeg -o cam2ip -ldflags "-s -w"


FROM scratch

COPY --from=build /build/cam2ip /cam2ip

EXPOSE 56000

ENTRYPOINT ["/cam2ip"]
