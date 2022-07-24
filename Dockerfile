FROM golang:1.18.0-alpine AS build-env
ENV GOOS=linux GOARCH=amd64 CGO_ENABLED=0
COPY . /build
WORKDIR /build
RUN go build -o prometheus_alert_overrider main.go

FROM gcr.io/distroless/static
WORKDIR /app
COPY --from=build-env /build/prometheus_alert_overrider /app
ENTRYPOINT ["./prometheus_alert_overrider"]