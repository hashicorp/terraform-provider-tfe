 #!/usr/bin/env bash
# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

LATEST=$(curl --silent "https://api.github.com/repos/hashicorp/terraform-plugin-docs/releases/latest" | jq -r .tag_name)
LATEST="${LATEST:1}"

curl -L https://github.com/hashicorp/terraform-plugin-docs/releases/download/v${LATEST}/tfplugindocs_${LATEST}_darwin_amd64.zip --output ${HOME}/Downloads/tfplugindocs.zip

unzip -u ${HOME}/Downloads/tfplugindocs.zip -d ${HOME}/Downloads/tfplugindir

cp ${HOME}/Downloads/tfplugindir/tfplugindocs ${HOME}/go/bin/tfplugindocs

rm -r ${HOME}/Downloads/tfplugindir ${HOME}/Downloads/tfplugindocs.zip

chmod 755 ${HOME}/go/bin/tfplugindocs

${HOME}/go/bin/tfplugindocs --help

# chmod 755 /usr/local/bin/tfplugindocsdir
#     /usr/local/bin/tfplugindocsdir/tfplugindocs


# curl -sSL -o tfplugindocs-darwin-amd64 https://github.com/hashicorp/terraform-plugin-docs/releases/latest/download/tfplugindocs-amd64
# sudo install -m 555 tfplugindocs-darwin-amd64 $HOME/Downloads
# rm tfplugindocs-darwin-amd64

#  curl -L https://github.com/hashicorp/terraform-plugin-docs/releases/download/v0.13.0/tfplugindocs_0.13.0_darwin_arm64.zip  --output /usr/local/bin/tfplugindocs.zip
#     unzip -u /usr/local/bin/tfplugindocs.zip -d /usr/local/bin/tfplugindocsdir
#     sudo chmod 755 /usr/local/bin/tfplugindocsdir
#     /usr/local/bin/tfplugindocsdir/tfplugindocs