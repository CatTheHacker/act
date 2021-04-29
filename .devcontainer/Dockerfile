# See here for image contents: https://github.com/microsoft/vscode-dev-containers/tree/v0.163.1/containers/go/.devcontainer/base.Dockerfile
# [Choice] Go version: 1, 1.16, 1.15, any other go version
ARG VARIANT=1.16.2
FROM golang:${VARIANT}

# [Option] Install zsh
ARG INSTALL_ZSH="true"
# [Option] Upgrade OS packages to their latest versions
ARG UPGRADE_PACKAGES="true"

# Install needed packages and setup non-root user. Use a separate RUN statement to add your own dependencies.
ARG USERNAME=vscode
ARG USER_UID=1000
ARG USER_GID=$USER_UID
ADD https://github.com/microsoft/vscode-dev-containers/raw/master/containers/go/.devcontainer/library-scripts/common-debian.sh /tmp/library-scripts/
ADD https://github.com/microsoft/vscode-dev-containers/raw/master/containers/go/.devcontainer/library-scripts/go-debian.sh /tmp/library-scripts/
ADD https://github.com/microsoft/vscode-dev-containers/raw/master/containers/go/.devcontainer/library-scripts/node-debian.sh /tmp/library-scripts/
ADD https://github.com/microsoft/vscode-dev-containers/raw/master/script-library/fish-debian.sh /tmp/library-scripts/
ADD https://github.com/microsoft/vscode-dev-containers/raw/master/script-library/docker-in-docker-debian.sh /tmp/library-scripts/
RUN bash /tmp/library-scripts/common-debian.sh "${INSTALL_ZSH}" "${USERNAME}" "${USER_UID}" "${USER_GID}" "${UPGRADE_PACKAGES}" \
    && apt-get clean -y && rm -rf /var/lib/apt/lists/*

# Install Go tools

RUN bash /tmp/library-scripts/go-debian.sh "none" "/usr/local/go" "${GOPATH}" "${USERNAME}" "false" \
    && apt-get clean -y

# [Option] Install Node.js
ARG INSTALL_NODE="true"
ARG NODE_VERSION="none"
ENV NVM_DIR=/usr/local/share/nvm
ENV NVM_SYMLINK_CURRENT=true \
    PATH=${NVM_DIR}/current/bin:${PATH}
RUN if [ "$INSTALL_NODE" = "true" ]; then bash /tmp/library-scripts/node-debian.sh "${NVM_DIR}" "${NODE_VERSION}" "${USERNAME}"; fi \
    && apt-get clean -y && rm -rf /var/lib/apt/lists/*

# [Optional] Uncomment this section to install additional OS packages.
# RUN apt-get update && export DEBIAN_FRONTEND=noninteractive \
#    && apt-get -y install --no-install-recommends fish

RUN bash /tmp/library-scripts/fish-debian.sh "true" "${USERNAME}" \
    && apt-get clean -y

RUN bash /tmp/library-scripts/docker-in-docker-debian.sh "true" "${USERNAME}" "true" \
    && apt-get clean -y

# [Optional] Uncomment the next line to use go get to install anything else you need
RUN go install github.com/jondot/goweight@latest \
    && go install github.com/junegunn/fzf@latest \
    && rm -rf /go/pkg

# [Optional] Uncomment this line to install global node packages.
# RUN su vscode -c "source /usr/local/share/nvm/nvm.sh && npm install -g <your-package-here>" 2>&1

RUN echo "#!/bin/sh\n\
    sudoIf() { if [ \"\$(id -u)\" -ne 0 ]; then sudo \"\$@\"; else \"\$@\"; fi }\n\
    SOCKET_GID=\$(stat -c '%g' /var/run/docker.sock) \n\
    if [ \"${SOCKET_GID}\" != '0' ]; then\n\
    if [ \"\$(cat /etc/group | grep :\${SOCKET_GID}:)\" = '' ]; then sudoIf groupadd --gid \${SOCKET_GID} docker-host; fi \n\
    if [ \"\$(id ${USERNAME} | grep -E \"groups=.*(=|,)\${SOCKET_GID}\(\")\" = '' ]; then sudoIf usermod -aG \${SOCKET_GID} ${USERNAME}; fi\n\
    fi\n\
    exec \"\$@\"" > /usr/local/share/docker-init.sh \
    && chmod +x /usr/local/share/docker-init.sh

# VS Code overrides ENTRYPOINT and CMD when executing `docker run` by default.
# Setting the ENTRYPOINT to docker-init.sh will configure non-root access to
# the Docker socket if "overrideCommand": false is set in devcontainer.json.
# The script will also execute CMD if you need to alter startup behaviors.
ENTRYPOINT [ "/usr/local/share/docker-init.sh" ]
CMD [ "sleep", "infinity" ]
