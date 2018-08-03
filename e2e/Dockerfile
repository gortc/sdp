FROM golang:1.10.3
ADD . /go/src/github.com/gortc/sdp
WORKDIR /go/src/github.com/gortc/sdp/e2e/
COPY e2e/main.go .
RUN go build .

FROM yukinying/chrome-headless-browser
COPY --from=0 /go/src/github.com/gortc/sdp/e2e/e2e .
COPY e2e/static static
CMD ["./e2e", "/usr/bin/google-chrome-unstable"]
ENTRYPOINT ["./e2e"]
