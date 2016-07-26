GROUPNAME := $(shell cat groupname.out)
GROUPFILE := "groupname.out"
TIMESTAMP := $(shell date +'%s') 

default: clean walinuxagent
	echo editions-$(TIMESTAMP) >$(GROUPFILE)
	azure group create --name editions-$(TIMESTAMP) --location centralus --template-file editions.json

clean:
	azure group delete $(GROUPNAME) -q --nowait || true
	rm -f $(GROUPFILE)

walinuxagent:
	make -C dockerfiles/walinuxagent push
