#!/bin/bash

echo 
echo
read -p "Resource and attribute descriptions are defined within the \"tfe\" directory resource schemas."
echo
read -p "Here is an example from tfe/resource_tfe_agent_pool: " 
echo 
echo "***** start *****"
echo
cat "4-descinfoex.sh"
echo
echo
echo "***** end *****"
echo
read -p "[1] A \"Description\" can be added for each attribute"
read -p "[2] An additional note modal can be added using the \"\n\n ~>\" syntax" 
