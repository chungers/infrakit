GROUPNAME := $(shell cat groupname.out)
GROUPFILE := "groupname.out"
CHECKSUM := $(shell sha1sum editions.json | awk '{ print $$1; }') 

default: clean
	echo editions-$(CHECKSUM) >$(GROUPFILE)
	azure group create --name editions-$(CHECKSUM) --location centralus --template-file editions.json

clean:
	azure group delete $(GROUPNAME) -q --nowait || true
	rm -f $(GROUPFILE)
