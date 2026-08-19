ARG OSRM_VERSION=v26.8.0-debian
FROM ghcr.io/project-osrm/osrm-backend:${OSRM_VERSION}

USER root
RUN apt-get update \
	&& apt-get install -y --no-install-recommends curl ca-certificates \
	&& rm -rf /var/lib/apt/lists/*

COPY entrypoint.sh /usr/local/bin/osrm-entrypoint.sh
RUN chmod +x /usr/local/bin/osrm-entrypoint.sh

WORKDIR /data
ENTRYPOINT ["/usr/local/bin/osrm-entrypoint.sh"]
