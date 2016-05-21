FROM scratch
ADD srvaddr.linux.x64 /srvaddr
ENV PATH=/
ENTRYPOINT ["srvaddr"]
