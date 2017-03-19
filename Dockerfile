FROM alpine:latest
MAINTAINER Łukasz Kurowski <crackcomm@gmail.com>

# Install certificates
RUN apk --update add ca-certificates

# Copy application
COPY ./dist/google /google

#
# Environment variables
# for google search app
#
ENV TOPIC google_search
ENV CHANNEL consumer
ENV NSQ_ADDR nsq:4150
ENV NSQLOOKUP_ADDR nsqlookup:4161
ENV CONCURRENCY 10

ENTRYPOINT ["/google"]
