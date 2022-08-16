#!/bin/bash
echo
read -p "A note can be added to any schema description using the \"\n\n ~> \" syntax."
echo
read -p "Here is an example from tfe/resource_tfe_agent_pool.go "
echo 
echo "***** start *****"
echo
cat "11-noteinfoex.sh"
echo
echo
read -p "***** end *****"
echo
echo "Visit https://registry.terraform.io/providers/hashicorp/tfe/latest/docs/resources/agent_pool to see the note on the website."
echo
read -p "[1] \"->\" will generate a note, while \"~>\" will generate a warning."