FROM golang:1-alpine as builder
WORKDIR /app
COPY go.mod go.sum .
RUN go mod download
COPY . .
RUN go build .

FROM gcr.io/distroless/static
WORKDIR /web
EXPOSE 8080
VOLUME /web
COPY --from=builder /app/ran /ran
ENTRYPOINT [ "/ran" ]
