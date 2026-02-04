FROM registry.suse.com/bci/golang:1.25

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
ARG RKE_PROVIDER_VERSION
ARG RANCHER2_PROVIDER_VERSION
ARG LOCALS_PROVIDER_VERSION
ARG CLOUD_PROVIDER_VERSION
ARG KUBERNETES_PROVIDER_VERSION
ARG LETS_ENCRYPT_EMAIL
ARG GOOGLE_TFP_EMAIL
ARG GOOGLE_TFP_SERVICE_ACCOUNT

ENV QASE_TEST_RUN_ID=${QASE_TEST_RUN_ID}
ENV TERRAFORM_VERSION=${TERRAFORM_VERSION}
ENV RKE_PROVIDER_VERSION=${RKE_PROVIDER_VERSION}
ENV RANCHER2_PROVIDER_VERSION=${RANCHER2_PROVIDER_VERSION}
ENV LOCALS_PROVIDER_VERSION=${LOCALS_PROVIDER_VERSION}
ENV CLOUD_PROVIDER_VERSION=${CLOUD_PROVIDER_VERSION}
ENV KUBERNETES_PROVIDER_VERSION=${KUBERNETES_PROVIDER_VERSION}
ENV LETS_ENCRYPT_EMAIL=${LETS_ENCRYPT_EMAIL}

RUN zypper install -y openssh wget unzip > /dev/null
RUN (curl -O https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/google-cloud-cli-linux-x86_64.tar.gz && \
    tar -xf google-cloud-cli-linux-x86_64.tar.gz && \
    ./google-cloud-sdk/install.sh --quiet && \
    rm google-cloud-cli-linux-x86_64.tar.gz) > /dev/null 2>&1

RUN zypper install -y python3 > /dev/null 2>&1
RUN zypper install -y python3-pip > /dev/null 2>&1
RUN zypper install -y azure-cli > /dev/null 2>&1

ENV PATH $PATH:/root/go/src/github.com/rancher/tfp-automation/google-cloud-sdk/bin

RUN (wget https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip && \
    zypper --non-interactive update && \
    unzip terraform_${TERRAFORM_VERSION}_linux_amd64.zip && \
    rm terraform_${TERRAFORM_VERSION}_linux_amd64.zip && \
    chmod a+x terraform && mv terraform /usr/local/bin/terraform) > /dev/null 2>&1

ARG CONFIG_FILE
COPY $CONFIG_FILE /config.yml

RUN mkdir /root/.ssh && chmod 600 .ssh/jenkins-*
RUN for pem_file in .ssh/jenkins-*; do \
        base_name="$(basename "$pem_file" .pem)"; \
        ssh-keygen -f "$pem_file" -y > ".ssh/${base_name}.pub"; \
    done

RUN if [[ "$RANCHER2_PROVIDER_VERSION" == *"-rc"* ]]; then \
      chmod +x ./scripts/setup-provider.sh && ./scripts/setup-provider.sh rancher2 v${RANCHER2_PROVIDER_VERSION} ; \
    fi;

RUN if [[ "$RKE_PROVIDER_VERSION" == *"-rc"* ]]; then \
      chmod +x ./scripts/setup-provider.sh && ./scripts/setup-provider.sh rke v${RKE_PROVIDER_VERSION} ; \
    fi;