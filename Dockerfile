FROM golang:1.24.5-alpine3.22 AS build


WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download


COPY . .
RUN go build -o main ./cmd


FROM alpine:3.22
WORKDIR /app
COPY --from=build /app/main .
# COPY .env .

CMD [ "/app/main" ]
EXPOSE 8484

