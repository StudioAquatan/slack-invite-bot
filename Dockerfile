FROM golang:1.11.4-alpine3.8

ENV PROJECT_DIR /go/src/slack-invite-bot

COPY ./app ${PROJECT_DIR}/app
COPY Gopkg.toml ${PROJECT_DIR}/Gopkg.toml
COPY Gopkg.lock ${PROJECT_DIR}/Gopkg.lock

WORKDIR ${PROJECT_DIR}

RUN apk update \
  && apk add --no-cache git \
  && go get -u github.com/golang/dep/cmd/dep \
#  && dep ensure

#RUN go build -i -o ./.build/slack-invite-bot

#CMD [".build/slack-invite-bot"]