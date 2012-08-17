
SRCS = true.go false.go yes.go md5sum.go sha1sum.go sha256sum.go sha224sum.go
LIBS = common.go
OBJS = $(SRCS:.go=)

all: $(SRCS)

%.go: dummy
	@echo " BUILD " $@
	@go build $@ $(LIBS)

clean:
	rm -f $(OBJS)

dummy:
	@true