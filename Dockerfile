FROM golang:1.16 as go-builder

ENV CGO_ENABLED 0

WORKDIR /workspace

COPY pkg/    pkg/
COPY main.go go.mod go.sum ./

RUN go build -o prometheus-cloudwatch-adapter main.go

FROM gcr.io/distroless/base

COPY --from=go-builder /workspace/prometheus-cloudwatch-adapter /prometheus-cloudwatch-adapter

EXPOSE 9513
USER 65534

ENTRYPOINT ["/prometheus-cloudwatch-adapter"]
