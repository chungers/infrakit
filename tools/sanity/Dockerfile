FROM python:2.7-alpine
COPY sanity.py /usr/bin
RUN chmod 755 /usr/bin/sanity.py
VOLUME /data
WORKDIR /data
ENTRYPOINT ["sanity.py"]
