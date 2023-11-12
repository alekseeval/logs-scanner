FROM ubuntu:latest
LABEL authors="reserv"

ENTRYPOINT ["top", "-b"]