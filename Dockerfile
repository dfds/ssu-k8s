FROM golang:1.24-alpine AS build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

COPY core /app/core
COPY feats /app/feats
COPY cmds /app/cmds

RUN go build -o /app/app /app/cmds/main.go

FROM golang:1.24-alpine

COPY --from=build /app/app /app/app

CMD [ "/app/app" ]