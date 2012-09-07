
SRCS = true.go false.go yes.go md5sum.go sha1sum.go sha256sum.go sha224sum.go cat.go tty.go tee.go pwd.go
LIBS = common.go
OBJS = $(SRCS:.go=)

all: $(OBJS)

%: %.go $(LIBS)
	@echo "  GO " $<
	@go build $< $(LIBS)

clean:
	rm -f $(OBJS)
