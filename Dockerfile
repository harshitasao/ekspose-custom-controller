FROM alpine
COPY ./ekspose-custom-controller /usr/loca/bin/ekspose
ENTRYPOINT [ "/usr/loca/bin/ekspose" ]