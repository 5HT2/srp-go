FROM golang:1.16.5

RUN mkdir /srp-go
ADD . /srp-go
RUN mkdir /srp-go/www/images
WORKDIR /srp-go

RUN go build -o srp-bin .

ENV ADDRESS "localhost:6060"
ENV MAXBODYSIZE "104857600"
CMD /srp-go/srp-bin -maxbodysize $MAXBODYSIZE -addr $ADDRESS
