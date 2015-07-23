FROM quay.io/geonet/golang-godep:latest

COPY . /go/src/github.com/GeoNet/geonet-news-md

WORKDIR /go/src/github.com/GeoNet/geonet-news-md

RUN godep go install -a

EXPOSE 8080

CMD ["/go/bin/geonet-news-md"]