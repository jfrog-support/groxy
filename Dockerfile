FROM golang:1.6-onbuild

RUN mkdir /root/.groxy/
ADD ./config.json /root/.groxy/ 
