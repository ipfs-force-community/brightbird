FROM ubuntu:20.04

# copy the app from build env
COPY  ./dist/testrunner /root/testrunner

ENTRYPOINT ["/root/testrunner"]