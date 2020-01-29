FROM golang:1-alpine as builder
WORKDIR /app
ADD . .
RUN go build .

FROM alpine
COPY --from=builder /app/ran /ran
WORKDIR /web
EXPOSE 8080
VOLUME /web
ENTRYPOINT [ "/ran" ]
