FROM alpine:3.5

RUN apk add --update bash ca-certificates jq groff less openssl cifs-utils openssl-dev libffi-dev python python-dev musl-dev linux-headers py-pip gcc py-setuptools \
		&& pip install --upgrade pip \
		&& pip install msrest \
		&& pip install msrestazure \
		&& pip install --pre azure \
		&& pip install pyparsing

ADD files/server.py /server.py
RUN chmod 755 /server.py

CMD ["python", "/server.py"]
