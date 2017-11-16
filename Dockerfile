FROM ubuntu:16.04

COPY client launch.sh /

RUN chmod +x /launch.sh

RUN chmod +x /client

CMD /launch.sh

