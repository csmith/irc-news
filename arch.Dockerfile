FROM golang:1.17 AS build
WORKDIR /app
COPY . /app
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -o /go/bin/app ./cmd/arch

FROM gcr.io/distroless/base-debian10
COPY --from=build /go/bin/app /
CMD ["/app"]
