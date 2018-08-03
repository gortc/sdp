FROM golang:1.10.3
WORKDIR /go/src/github.com/gortc/sdp/e2e/
ADD vendor /go/src/github.com/gortc/sdp/e2e/vendor
COPY main.go .
RUN go build .

FROM yukinying/chrome-headless-browser
COPY --from=0 /go/src/github.com/gortc/sdp/e2e/e2e .
CMD ["./e2e"]
ENTRYPOINT ["./e2e"]
