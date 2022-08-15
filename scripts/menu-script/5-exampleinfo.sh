#!/bin/bash
echo
echo
read -p "Add code snippets to an example file, demonstrating how your resource is declared in a terraform config file with blocks, arguments, and expressions."
echo
read -p "Here is an example from examples/resource_tfe_agent_pool/resource.tf: " 
echo 
echo "***** start *****"
echo
cat "6-exampleinfoex.sh"
echo
echo
read -p "***** end *****"
echo
while true; do
    echo
    read -p "-> Generate an empty example file for your resource? [y|n] " yn
    case $yn in
        [Yy]* ) 
        ./7-createexfile.sh
        echo
        break;;
        [Nn]* ) 
        echo
        read -p "File not created. Examples should be added to a file with the path examples/resource_directory/example.tf" ; exit;;
        * ) echo; echo "Please answer yes or no.";;
    esac
done
read -p "If you require more than one example file go to \"Main Menu\" and then \"Custom\"".
