FROM golang:1.24-alpine AS build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

COPY core /app/core
COPY feats /app/feats
COPY cmds /app/cmds

RUN go build -o /app/app /app/cmds/main.go

FROM golang:1.24-alpine

WORKDIR /app

RUN apk update && apk add git openssh

RUN adduser \
  --disabled-password \
  --home /app \
  --gecos '' app \
  && chown -R app /app
USER app

COPY --chown=app:app static/known_hosts /app/.ssh/known_hosts
COPY --chown=app:app static/gitconfig /app/.ssh/config

COPY --chown=app:app --from=build /app/app /app/app
COPY --chown=app:app core /app/core

CMD [ "/app/app" ]