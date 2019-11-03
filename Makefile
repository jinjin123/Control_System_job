# Go parameters
goCmd=go
goBuild=$(goCmd) build
goClean=$(goCmd) clean
goTest=$(goCmd) test
goGet=$(goCmd) get
sourceAdmDir=./app/jiacrontab_admin
sourceNodeDir=./app/jiacrontabd
sourceNodeODir=./app/jiacrontabo
binAdm=$(sourceAdmDir)/jiacrontab_admin
binNode=$(sourceNodeDir)/jiacrontabd
binNodeO=$(sourceNodeODir)/jiacrontabo

buildDir=./build
buildAdmDir=$(buildDir)/jiacrontab/jiacrontab_admin
buildNodeDir=$(buildDir)/jiacrontab/jiacrontabd
buildNodeODir=$(buildDir)/jiacrontab/jiacrontabdo

admCfg=$(sourceAdmDir)/jiacrontab_admin.ini
nodeCfg=$(sourceNodeDir)/jiacrontabd.ini
nodeOCfg=$(sourceNodeODir)/jiacrontabd.ini
staticDir=./jiacrontab_admin/static/build
staticSourceDir=./jiacrontab_admin/static
workDir=$(shell pwd)


.PHONY: all build test clean build-linux build-windows
all: test build
build:
	$(call init)
	$(goBuild) -mod=vendor -o $(binAdm) -v $(sourceAdmDir)
	$(goBuild) -mod=vendor -o $(binNode) -v $(sourceNodeDir)
	$(goBuild) -mod=vendor -o $(binNodeO) -v $(sourceNodeODir)
	mv $(binAdm) $(buildAdmDir)
	mv $(binNode) $(buildNodeDir)
	mv $(binNodeO) $(buildNodeODir)
test:
	$(goTest) -mod=vendor -v -race -coverprofile=coverage.txt -covermode=atomic $(sourceAdmDir)
	$(goTest) -mod=vendor -v -race -coverprofile=coverage.txt -covermode=atomic $(sourceNodeDir)
	$(goTest) -mod=vendor -v -race -coverprofile=coverage.txt -covermode=atomic $(sourceNodeODir)
clean:
	rm -f $(binAdm)
	rm -f $(binNode)
	rm -f $(binNodeO)
	rm -rf $(buildDir)


# Cross compilation
build-linux:
	$(call init)
	GOOS=linux GOARCH=amd64 $(goBuild) -mod=vendor -o $(binAdm) -v $(sourceAdmDir)
	GOOS=linux GOARCH=amd64 $(goBuild) -mod=vendor -o $(binNode) -v $(sourceNodeDir)
	GOOS=linux GOARCH=amd64 $(goBuild) -mod=vendor -o $(binNodeO) -v $(sourceNodeODir)
	mv $(binAdm) $(buildAdmDir)
	mv $(binNode) $(buildNodeDir)
	mv $(binNodeO) $(buildNodeODir)

build-windows:
	$(call init)
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC="x86_64-w64-mingw32-gcc -fno-stack-protector -D_FORTIFY_SOURCE=0 -lssp" $(goBuild) -mod=vendor -o $(binAdm).exe -v $(sourceAdmDir)
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC="x86_64-w64-mingw32-gcc -fno-stack-protector -D_FORTIFY_SOURCE=0 -lssp" $(goBuild) -mod=vendor -o $(binNode).exe -v $(sourceNodeDir)

	mv $(binAdm).exe $(buildAdmDir)
	mv $(binNode).exe $(buildNodeDir)


define init
	rm -rf $(buildDir)
	mkdir $(buildDir)
	mkdir -p $(buildAdmDir)
	mkdir -p $(buildNodeDir)
	mkdir -p $(buildNodeODir)

	cp $(admCfg) $(buildAdmDir)
	cp $(nodeCfg) $(buildNodeDir)
	cp $(nodeOCfg) $(buildNodeODir)
endef
