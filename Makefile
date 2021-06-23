# Makefile for go app
# author: julien-beguier

SRCSERVER=	cmd/gs-server
SRCCLIENT=	cmd/gs-client
BIN=		bin

GOBUILD=	go build
RM=		rm -f
GSSERVER=	gs-server
GSCLIENT=	gs-client

all: server client

server:
	$(GOBUILD) -o $(BIN)/$(GSSERVER) $(SRCSERVER)/server.go
	@echo "\n\033[1;31mBuild $(GSSERVER) complete\033[0;0m\n"

client:
	$(GOBUILD) -o $(BIN)/$(GSCLIENT) $(SRCCLIENT)/client.go
	@echo "\n\033[1;31mBuild $(GSCLIENT) complete\033[0;0m\n"

fclean:
	$(RM) $(BIN)/$(GSSERVER) $(BIN)/$(GSCLIENT)
	@echo "\n\033[1;31mRemoved \033[1;33m$(BIN)/$(GSSERVER) & $(BIN)/$(GSCLIENT)\033[0;0m\n"

re:	fclean server client

.PHONY:	all server client fclean re
