FROM golang:1.23-alpine AS build
WORKDIR /build
ADD . .
RUN CGO_ENABLED=0 go build -o kudos .

FROM scratch
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /build/kudos .
ENTRYPOINT ["./kudos"]
