FROM golang:1.22

ENV GOPATH /root/go
ENV PATH ${PATH}:/root/go/bin

ENV WORKSPACE ${GOPATH}/src/github.com/rancher/tfp-automation

WORKDIR $WORKSPACE

COPY . ./
SHELL ["/bin/bash", "-c"]

RUN go mod download && \
    go install gotest.tools/gotestsum@latest

ARG QASE_TEST_RUN_ID
ARG TERRAFORM_VERSION
ARG EXTERNAL_ENCODED_VPN
ARG VPN_ENCODED_LOGIN
ARG RKE_PROVIDER_VERSION
ARG RANCHER2_PROVIDER_VERSION
ARG LOCALS_PROVIDER_VERSION
ARG AWS_PROVIDER_VERSION
ARG RANCHER2_KEY_PATH
ARG RKE_KEY_PATH
ARG SANITY_KEY_PATH

ENV QASE_TEST_RUN_ID=${QASE_TEST_RUN_ID}
ENV TERRAFORM_VERSION=${TERRAFORM_VERSION}
ENV EXTERNAL_ENCODED_VPN=${EXTERNAL_ENCODED_VPN}
ENV VPN_ENCODED_LOGIN=${VPN_ENCODED_LOGIN}
ENV RKE_PROVIDER_VERSION=${RKE_PROVIDER_VERSION}
ENV RANCHER2_PROVIDER_VERSION=${RANCHER2_PROVIDER_VERSION}
ENV LOCALS_PROVIDER_VERSION=${LOCALS_PROVIDER_VERSION}
ENV AWS_PROVIDER_VERSION=${AWS_PROVIDER_VERSION}
ENV RANCHER2_KEY_PATH=${RANCHER2_KEY_PATH}
ENV RKE_KEY_PATH=${RKE_KEY_PATH}
ENV SANITY_KEY_PATH=${SANITY_KEY_PATH}

RUN wget https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip -q && apt-get update > /dev/null && apt-get install unzip > /dev/null && \
    unzip terraform_${TERRAFORM_VERSION}_linux_amd64.zip > /dev/null && \ 
    rm terraform_${TERRAFORM_VERSION}_linux_amd64.zip > /dev/null && \
    chmod a+x terraform > /dev/null && mv terraform /usr/local/bin/terraform > /dev/null

ARG CONFIG_FILE
COPY $CONFIG_FILE /config.yml

ARG PEM_FILE
COPY $PEM_FILE /key.pem
RUN echo $PEM_FILE > key.pem && chmod 600 key.pem

RUN if [[ -z '$EXTERNAL_ENCODED_VPN' ]] ; then \
      echo 'no vpn provided' ; \
    else \
      apt-get update > /dev/null && apt-get -y install sudo openvpn net-tools > /dev/null; \
    fi;

RUN if [[ "$RANCHER2_PROVIDER_VERSION" == *"-rc"* ]]; then \
      chmod +x ./scripts/setup-provider.sh && ./scripts/setup-provider.sh rancher2 v${RANCHER2_PROVIDER_VERSION} ; \
    fi;

RUN if [[ "$RKE_PROVIDER_VERSION" == *"-rc"* ]]; then \
      chmod +x ./scripts/setup-provider.sh && ./scripts/setup-provider.sh rke v${RKE_PROVIDER_VERSION} ; \
    fi;