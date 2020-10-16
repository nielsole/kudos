FROM golang:1.15-alpine as build
WORKDIR /build
ADD . .
RUN CGO_ENABLED=0 go build -o kudos .
FROM scratch
COPY --from=build /build/kudos .
ENTRYPOINT ["./kudos"]
