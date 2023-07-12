FROM alpine:latest

EXPOSE 8001

COPY ./main /opt

CMD ["/opt/main"]

