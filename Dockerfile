FROM golang:1.19

# USER root

# Create a new user 'myuser'
RUN useradd -m myuser
RUN usermod -aG sudo myuser
RUN usermod -aG docker myuser
USER myuser
WORKDIR /home/myuser


RUN apt-get update && apt-get install -y sudo

RUN mkdir -p /.cache && chmod -R 777 /.cache

RUN mkdir -p $GOPATH/pkg/mod && chmod -R 777 $GOPATH/pkg/mod

RUN chown -R myuser:sudo $GOPATH/pkg/mod && chmod -R g+rwx $GOPATH/pkg/mod

# WORKDIR /usr/app/src

COPY [".", "$WORKDIR"]

ADD ./* ./
SHELL ["/bin/bash", "-c"] 

RUN go mod download && \
    go install gotest.tools/gotestsum@latest

ARG TERRAFORM_VERSION=1.6.5
RUN wget https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip && apt-get update && apt-get install unzip &&  unzip terraform_${TERRAFORM_VERSION}_linux_amd64.zip && rm terraform_${TERRAFORM_VERSION}_linux_amd64.zip && chmod u+x terraform && mv terraform /usr/bin/terraform

RUN apt-get update -y && apt-get install -y gzip

ARG CONFIG_FILE
COPY ${CONFIG_FILE} /config.yml

ARG RANCHER2_PROVIDER_VERSION
RUN chmod +x scripts/setup-provider.sh && ./scripts/setup-provider.sh rancher2 v${RANCHER2_PROVIDER_VERSION}

# remove sudo access from myuser
RUN gpasswd -d myuser sudo





