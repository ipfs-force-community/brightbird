FROM ubuntu:20.04

# copy the app from build env
COPY  ./testrunner /root/testrunner

ENTRYPOINT ["/root/testrunner"]