##############################################################
## Stage 1 - Go Build As builder
##############################################################

FROM golang:1.23 AS builder
WORKDIR /opt
COPY . .
RUN go build -o app .

#############################################################
## Stage 2 - Application Setup AS prod
##############################################################
FROM ubuntu:22.04 AS prod
ARG UID=0
ARG GID=0
ARG USER=root
ARG GROUP=root
# #-------------------------------------------------------------
RUN apt-get update -y && apt-get upgrade -y  
RUN apt-get install -y  curl
# RUN apt-get install -y ca-certificates curl jq gnupg software-properties-common wget lsb-release && rm -rf /var/lib/apt/lists/*

# #-------------------------------------------------------------
# # Add the HashiCorp GPG key and official repository for Terraform
# RUN wget -O- https://apt.releases.hashicorp.com/gpg | gpg --dearmor | tee /usr/share/keyrings/hashicorp-archive-keyring.gpg && \
#     echo "deb [signed-by=/usr/share/keyrings/hashicorp-archive-keyring.gpg] https://apt.releases.hashicorp.com $(lsb_release -cs) main" | tee /etc/apt/sources.list.d/hashicorp.list && \
#     apt update && apt install -y terraform && \
#     rm -rf /var/lib/apt/lists/*
#-------------------------------------------------------------
# User Set
RUN if [ "${USER}" != "root" ]; then \
        groupadd -g ${GID} ${GROUP} && \
        useradd -m -u ${UID} -g ${GID} -s /bin/bash ${USER}; \
    fi
#-------------------------------------------------------------
# Copy App and Web
RUN mkdir -p /app/log
RUN chown -R ${USER}:${GROUP} /app
USER ${USER}
COPY --from=builder --chown=${USER}:${GROUP} /opt/app /app/app
COPY --from=builder --chown=${USER}:${GROUP} /opt/web /app/web
COPY --from=builder --chown=${USER}:${GROUP} /opt/data/var /app/data/var
COPY --from=builder --chown=${USER}:${GROUP} /opt/scripts /app/scripts
#-------------------------------------------------------------
# Add entrypoint script
COPY --chown=${USER}:${GROUP} entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh
#-------------------------------------------------------------
WORKDIR /app
EXPOSE 3300
ENTRYPOINT ["/app/entrypoint.sh"]
