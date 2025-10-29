# syntax=docker/dockerfile:1
FROM python:3.10-slim-buster AS pythonbuilder
ENV PYTHONUNBUFFERED=1
RUN pip install requests
RUN pip install lxml
RUN mkdir /scripts
RUN mkdir /internal
RUN ls
COPY ./tachoparser/scripts/ /scripts/
COPY ./tachoparser/internal/ /internal/
WORKDIR /scripts/pks1
RUN ./dl_all_pks1.py
WORKDIR /scripts/pks2
RUN ./dl_all_pks2.py

FROM golang:1.24 AS gobuilder
RUN apt install libc6
WORKDIR /go/src/github.com/TaCtrlai/tachoparser
COPY ./ ./
COPY --from=pythonbuilder /internal/pkg/certificates/pks1/ internal/pkg/certificates/pks1/
COPY --from=pythonbuilder /internal/pkg/certificates/pks2/ internal/pkg/certificates/pks2/
RUN go mod vendor
WORKDIR /go/src/github.com/TaCtrlai/tachoparser/cmd/dddparser
RUN go build .
WORKDIR /go/src/github.com/TaCtrlai/tachoparser/cmd/dddserver
RUN go build .
WORKDIR /go/src/github.com/TaCtrlai/tachoparser/cmd/dddclient
RUN go build .

FROM ubuntu
RUN apt install libc6
COPY --from=gobuilder /etc/ssl/certs/* /etc/ssl/certs/
COPY --from=gobuilder /usr/share/zoneinfo/* /usr/share/zoneinfo/
COPY --from=gobuilder /go/src/github.com/TaCtrlai/tachoparser/cmd/dddserver/dddserver /dddserver
ENTRYPOINT ["/dddserver"]
CMD []