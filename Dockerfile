FROM alpine:3.6

COPY /build/app /bin/

ENTRYPOINT ["/bin/app"]
