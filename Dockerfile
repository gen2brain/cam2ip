FROM golang:alpine as build

WORKDIR /build

RUN apk add --no-cache git

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o cam2ip -trimpath -ldflags "-s -w" github.com/gen2brain/cam2ip/cmd/cam2ip


FROM scratch

COPY --from=build /build/cam2ip /cam2ip

EXPOSE 56000

ENTRYPOINT ["/cam2ip"]
