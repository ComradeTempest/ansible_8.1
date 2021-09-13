# Build cdrsender using golang image
FROM golang as builder

COPY . /go/src/tas_cdrsender
WORKDIR /go/src/tas_cdrsender
RUN go build .

# Run
FROM alpine

COPY --from=builder go/src/tas_cdrsender/cdrsender /cdrsender
COPY --from=builder go/src/tas_cdrsender/cdrsender.ini /cdrsender.ini
CMD ["/cdrsender"]
EXPOSE 2120
