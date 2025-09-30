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
FROM debian:12-slim AS prod

# User Set
ARG UID=0
ARG GID=0
ARG USER=root
ARG GROUP=root
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

WORKDIR /app
EXPOSE 3300

ENTRYPOINT ["/app/app"]