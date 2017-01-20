SUBDIRS := $(wildcard */)

.PHONY: all $(SUBDIRS)
all: $(SUBDIRS)

$(SUBDIRS):
	make -C $@
