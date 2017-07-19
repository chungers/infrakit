FROM alpine:3.6

MAINTAINER French Ben <FrenchBen@Docker.com>

RUN apk add --update bash ca-certificates jq groff less openssl-dev libffi-dev python python-dev musl-dev linux-headers py-pip gcc py-setuptools \
    && pip install msrest \
    && pip install msrestazure \
    && pip install --pre azure \
    && pip install docker-py \
    && pip install pytz \
    && apk --purge -v del py-pip \
    && rm -rf /var/cache/apk/*

ADD files/azlookup.py /usr/bin/azlookup
ADD files/azendpt.py /usr/bin
ADD files/aztags.py /usr/bin
RUN chmod +x /usr/bin/azlookup

ENTRYPOINT ["azlookup"]
CMD ["--help"]
