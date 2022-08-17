 curl -L https://github.com/hashicorp/terraform-plugin-docs/releases/download/v0.13.0/tfplugindocs_0.13.0_darwin_arm64.zip  --output /usr/local/bin/tfplugindocs.zip
    unzip -u /usr/local/bin/tfplugindocs.zip -d /usr/local/bin/tfplugindocsdir
    sudo chmod 755 /usr/local/bin/tfplugindocsdir
    /usr/local/bin/tfplugindocsdir/tfplugindocs