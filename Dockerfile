FROM node:13.7.0-buster-slim as builder
RUN echo 'Acquire::http::Proxy "http://proxies.labs:3142/apt-cacher/";' > /etc/apt/apt.conf.d/01proxy
RUN apt-get update && apt-get install --no-install-recommends -y git openssh-client ca-certificates python2 make g++ 
RUN apt-get install -y --no-install-recommends s4cmd curl

ADD package-lock.json /tmp/package-lock.json
ADD package.json /tmp/package.json
RUN cd /tmp && npm install
RUN mkdir -p /usr/src 

COPY . /usr/src/app/
WORKDIR /usr/src/app/

RUN ln -s /tmp/node_modules /usr/src/app/node_modules
ENTRYPOINT /bin/bash
RUN npm run build

ARG TAG
ARG S3_ACCESS_KEY
ARG S3_SECRET_KEY
ARG BUCKET_NAME
ARG REPO_NAME
ARG SUBPROJECT
ENV http_proxy=
RUN s4cmd --endpoint-url=http://ci.labs:9000 ls s3://${BUCKET_NAME}/${SUBPROJECT}/${TAG}/ || s4cmd --endpoint-url=http://ci.labs:9000 mb s3://${BUCKET_NAME}/${SUBPROJECT}/${TAG}/

ARG ARTIFACTS
RUN echo "About to upload artifacts.."
RUN for artifact in ${ARTIFACTS}; do \
	echo "Artifact: $artifact"; \
	if [ -d $artifact ]; then \
		cd $artifact && tar czf /tmp/${artifact%/}.tar.gz .; \
		artifact=/tmp/${artifact%/}.tar.gz; \
		echo "Artifact: $artifact"; \
	else \
		artifact=$PWD/$artifact; \
	fi;\
	s4cmd --endpoint-url=http://ci.labs:9000 put --force $artifact s3://${BUCKET_NAME}/${SUBPROJECT}/${TAG}/; \
	done
RUN curl -s http://david-dotopc.labs:8080/deploy/${REPO_NAME}/${SUBPROJECT}/${TAG}
