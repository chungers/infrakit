# FROM docs/base:latest
FROM appaws/docs-editions:latest
MAINTAINER Patrick Chanezon <patrick.chanezon@docker.com> (@chanezon)

ENV PROJECT=editions
# To get the git info for this repo

RUN cd /docs/content/;rm index.md
RUN cd /docs/layouts/;rm -rf mac
RUN cd /docs/layouts/;rm -rf windows
RUN apt-get update && apt-get install -y inotify-tools

COPY . /src

COPY . /docs
COPY dev.sh /bin/

# Sed to process GitHub Markdown
# 1-2 Remove comment code from metadata block
# 3 Change ](/word to ](/project/ in links
# 4 Change ](word.md) to ](/project/word)
# 5 Remove .md extension from link text
# 6 Change ](../ to ](/project/word)
# 7 Change ](../../ to ](/project/ --> not implemented
#
#
# RUN find /docs/content -type f -name "*.md" -exec sed -i.old \
#    -e '/^<!.*metadata]>/g' \
#    -e '/^<!.*end-metadata.*>/g' {} \;
