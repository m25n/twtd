FROM gcr.io/distroless/static
ARG TARGETPLATFORM

COPY $TARGETPLATFORM/twtd .

CMD ["/twtd"]