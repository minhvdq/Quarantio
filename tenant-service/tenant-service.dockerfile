FROM alpine:latest

RUN mkdir /app

COPY tenantApp /app

CMD ["/app/tenantApp"]
