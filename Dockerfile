FROM denismakogon/gocv-alpine:3.4.2-buildstage as build

WORKDIR /go/src/github.com/DanShu93/go-cv

COPY . .

RUN go get ./...
RUN go build -o $GOPATH/bin/server ./server.go

FROM denismakogon/gocv-alpine:3.4.2-runtime

COPY --from=build /go/bin/server /server

ENTRYPOINT ["/server"]
