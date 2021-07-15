FROM golang:1.16.5

RUN mkdir /srp-go
ADD . /srp-go
RUN mkdir /srp-go/www/images
WORKDIR /srp-go

RUN apt-get update && apt-get install -y \
    libvips \
 && rm -rf /var/lib/apt/lists/*

RUN go build -o srp-bin .

ENV ADDRESS "localhost:6060"
ENV MAXBODYSIZE "104857600"
CMD /srp-go/srp-bin -maxbodysize $MAXBODYSIZE -addr $ADDRESS
