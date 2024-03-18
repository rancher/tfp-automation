FROM golang:1.22

RUN mkdir -p /.cache && chmod -R 777 /.cache
RUN mkdir -p $GOPATH/pkg/mod && chmod -R 777 $GOPATH/pkg/mod
RUN chown -R root:root $GOPATH/pkg/mod && chmod -R g+rwx $GOPATH/pkg/mod
RUN groupadd -g 1000 jenkins && useradd -u 1000 -g jenkins -m jenkins

WORKDIR /usr/app/src

COPY [".", "$WORKDIR"]

ADD ./* ./
SHELL ["/bin/bash", "-c"]

RUN go mod download && \
    go install gotest.tools/gotestsum@latest

ARG TERRAFORM_VERSION
ARG EXTERNAL_ENCODED_VPN
ARG VPN_ENCODED_LOGIN
ARG RANCHER2_PROVIDER_VERSION

RUN wget https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip && apt-get update && apt-get install unzip && \
    unzip terraform_${TERRAFORM_VERSION}_linux_amd64.zip && \ 
    rm terraform_${TERRAFORM_VERSION}_linux_amd64.zip && \
    chmod a+x terraform && mv terraform /usr/local/bin/terraform

ARG CONFIG_FILE
COPY $CONFIG_FILE /config.yml

RUN if [[ -z '$EXTERNAL_ENCODED_VPN' ]] ; then \
      echo 'no vpn provided' ; \
    else \
      apt-get update && apt-get -y install sudo openvpn net-tools ; \
    fi;

RUN if [[ "$RANCHER2_PROVIDER_VERSION" == *"-rc"* ]]; then \
      chmod +x ./scripts/setup-provider.sh && ./scripts/setup-provider.sh rancher2 v${RANCHER2_PROVIDER_VERSION} ; \
    fi;

USER jenkins