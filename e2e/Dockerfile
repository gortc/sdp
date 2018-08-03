ARG CI_GO_VERSION
FROM golang:${CI_GO_VERSION}
ADD . /go/src/github.com/gortc/sdp
WORKDIR /go/src/github.com/gortc/sdp/e2e/
RUN go build .

FROM yukinying/chrome-headless-browser
COPY --from=0 /go/src/github.com/gortc/sdp/e2e/e2e .
COPY e2e/static static
ENTRYPOINT ["./e2e", "-b=/usr/bin/google-chrome-unstable", "-timeout=3s"]
