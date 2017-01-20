FROM google/cloud-sdk@sha256:fb904276e8a902ccd9564989d9222bdfbe37ffcd7f9989ca7e24b4019a9b4b6b

RUN apt-get update && apt-get install -y python-setuptools

RUN easy_install-2.7 Jinja2

WORKDIR /tests
ENTRYPOINT [ "/tests/files/runner.sh" ]
CMD [ "all" ]
COPY templates /templates
COPY tests /tests
