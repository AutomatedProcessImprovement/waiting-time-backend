FROM golang:1.18 AS wt_build
WORKDIR /go/src
COPY app ./app
COPY main.go .
COPY go.mod .
COPY go.sum .

ARG MOD_NAME=github.com/AutomatedProcessImprovement/waiting-time-backend
ARG OUT_BIN=wt_server

ENV CGO_ENABLED=0
RUN go get $MOD_NAME
RUN go build -a -installsuffix cgo -o $OUT_BIN $MOD_NAME

FROM scratch AS wt_runtime
COPY --from=wt_build /go/src/$OUT_BIN ./
EXPOSE 8080/tcp
ENTRYPOINT ["./$OUT_BIN"]
