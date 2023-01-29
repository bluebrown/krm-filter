FROM alpine

RUN apk --no-cache --update add bash git
SHELL ["/bin/bash", "-c"]

COPY bin /opt/krm-filter/bin
ENV PATH="/opt/krm-filter/bin:$PATH"

COPY e2e/cmp/build/plugin.yaml /home/argocd/cmp-server/config/plugin.yaml
COPY --chmod=0755 e2e/cmp/build/plugin.sh /opt/krm-filter/bin/plugin

COPY --from=gcr.io/kpt-dev/kpt:v1.0.0-beta.24 /usr/local/bin/kpt /usr/local/bin/kpt
COPY --from=gcr.io/kpt-fn/remove-local-config-resources:v0.1.0 /usr/local/bin/function /opt/krm-filter/bin/remove-local-config-resources

RUN adduser --system --disabled-password --gecos "" --shell /bin/bash --uid 999 argocd
USER 999
