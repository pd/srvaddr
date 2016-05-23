FROM scratch
ADD srvaddr_linux_amd64 /srvaddr
ENV PATH=/
ENTRYPOINT ["srvaddr"]
