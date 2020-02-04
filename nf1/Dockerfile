# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2019 Intel Corporation

FROM centos:7.6.1810

RUN yum install -y openssl \
    && yum clean all

RUN yum install -y net-tools tcpdump ethtool

WORKDIR /root/

COPY ./nf1 .
COPY nf.json ./config/
EXPOSE 8060:8060
EXPOSE 8070:8070
CMD ["./nf1"]
