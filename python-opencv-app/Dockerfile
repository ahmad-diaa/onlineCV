FROM lpetre/alpine-opencv-microimage:p3-3.2

ADD ca-certificates.crt /etc/ssl/certs/
ADD python-opencv-app /
ADD onlineCV-b3ad0190adc1.json /


CMD ["/python-opencv-app"]
EXPOSE 8000
