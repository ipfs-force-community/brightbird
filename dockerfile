FROM ubuntu:22.04

# copy the app from build env
COPY  ./testrunner /root/testrunner

ENTRYPOINT ["/root/testrunner"]