FROM golang:1.21 AS build-stage

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /jig

FROM build-stage AS run-test-stage
RUN go test -v ./...

FROM gcr.io/distroless/base-debian12 AS build-release-stage
WORKDIR /

COPY --from=build-stage /jig .

USER nonroot:nonroot
ENTRYPOINT ["./jig"]