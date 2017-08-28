FROM alpine:latest

ADD ca-certificates.crt /etc/ssl/certs/
ADD app /
ADD templates /templates
ADD private /private
ADD myconfig.json /
ADD onlineCV-b3ad0190adc1.json /
CMD ["/app","-configfile=myconfig.json"]
EXPOSE 3000
