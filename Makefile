# variables
GOCMD			=	go
GOPATH			:=	${shell pwd}
BINPATH			=	$(GOPATH)/bin

# parameters
GODEP			=	$(GOCMD) get
GOTEST			=	$(GOCMD) test -v 
GOBUILD			=	$(GOCMD) build
GOINSTALL		=	$(GOCMD) install

export GOPATH

# buildable packages
MAIN_PKGS 		:=	fbs.com/sheduler \
									fbs.com/sheduler/providers \
									fbs.com/sheduler/types

# usable libraries
LIBS_PKGS 		:=

# dependencies packages
DEPS_PKGS 		:=		gopkg.in/yaml.v2 \
										github.com/lib/pq \
										github.com/go-gorp/gorp \
										github.com/erikstmartin/go-testdb \
										github.com/smartystreets/goconvey



# packages for testing
TEST_PKGS		:=	fbs.com/sheduler \
								fbs.com/sheduler/providers \
								fbs.com/sheduler/types 


# buildable lists
DEPS_LIST		=	$(foreach int, $(DEPS_PKGS), $(int)_deps)
TEST_LIST		=	$(foreach int, $(TEST_PKGS), $(int)_test)
LIBS_LIST		=	$(foreach int, $(LIBS_PKGS), $(int)_libs)
BUILD_LIST		=	$(foreach int, $(MAIN_PKGS), $(int)_build)
INSTALL_LIST	=	$(foreach int, $(MAIN_PKGS), $(int)_install)

# targets
.PHONY:			$(DEPS_LIST) $(TEST_LIST) $(LIBS_LIST) $(BUILD_LIST) $(INSTALL_LIST)

all:			deps test libs build

deps:			$(DEPS_LIST)
test:			$(TEST_LIST)
libs:			$(LIBS_LIST)
build:			$(BUILD_LIST)
install:		$(INSTALL_LIST)

$(DEPS_LIST): %_deps:
	$(GODEP) $*

$(TEST_LIST): %_test:
	$(GOTEST) $*

$(LIBS_LIST): %_libs:
	$(GOBUILD) -o $(BINPATH)/$(shell basename $*) $*

$(BUILD_LIST): %_build:
	$(GOBUILD) -o $(BINPATH)/$(shell basename $*) $*

$(INSTALL_LIST): %_install:
	$(GOINSTALL) $*
