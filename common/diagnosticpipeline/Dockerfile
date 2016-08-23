FROM alpine

RUN apk add --update --no-cache python3 py-pip
RUN pip install awscli
COPY /bundle-diagnostics.sh /bundle-diagnostics.sh
ENTRYPOINT ["/bundle-diagnostics.sh"]
