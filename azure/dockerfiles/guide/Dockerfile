FROM docker4x/init-azure@sha256:0117699b93ab0e2ddde9a0dee1c06a1d7c579edd7fff7c1a29d09a677bffabde

RUN apk add --update bash ca-certificates jq groff less openssl-dev libffi-dev python python-dev musl-dev linux-headers py-pip gcc py-setuptools \
    && pip install msrest \
    && pip install msrestazure \
    && pip install --pre azure \
    && pip install docker-py \
    && pip install pytz \
    && apk --purge -v del py-pip \
    && rm -rf /var/cache/apk/*

# setup cron
RUN mkdir -p /usr/docker /var/log/docker/
ADD files/crontab.txt /usr/docker/crontab.txt
ADD files/buoy.sh /usr/bin
ADD files/refresh.sh /usr/bin
ADD bin/buoy /usr/bin
ADD files/vacuum.sh /usr/bin
ADD files/azparameters.py /usr/bin
ADD files/azutils.py /usr/bin
ADD files/azupg_listener_log_cfg.json /etc
ADD files/azupg_listener.py /usr/bin
ADD files/kick_azupg_listener.sh /usr/bin
ADD files/azrejoin_log_cfg.json /etc
ADD files/azrejoin.py /usr/bin
ADD files/kick_azrejoin_listener.sh /usr/bin
ADD files/aznodewatch_log_cfg.json /etc
ADD files/aznodewatch.py /usr/bin
ADD files/kick_aznodewatch.sh /usr/bin
ADD files/azendpt.py /usr/bin
COPY files/entry.sh /entry.sh
RUN chmod 755 /entry.sh \
    /usr/bin/azparameters.py /usr/bin/vacuum.sh \
    /usr/bin/buoy /usr/bin/buoy.sh /usr/bin/refresh.sh \
    /usr/bin/azutils.py \
    /usr/bin/azupg_listener.py /usr/bin/kick_azupg_listener.sh \
    /usr/bin/azrejoin.py /usr/bin/kick_azrejoin_listener.sh \
    /usr/bin/aznodewatch.py /usr/bin/kick_aznodewatch.sh

RUN /usr/bin/crontab /usr/docker/crontab.txt

CMD ["/entry.sh"]
