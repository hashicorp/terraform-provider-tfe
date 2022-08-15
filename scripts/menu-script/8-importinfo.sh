#!/bin/bash
echo
read -p "Add an import statement file to show how your resource can be imported (optional)."
echo
read -p "Here is an example from examples/resource_tfe_agent_pool/import.sh: " 
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
    read -p "-> Generate an empty import statement file? [y|n] " yn
    case $yn in
        [Yy]* ) 
        ./9-createimpfile.sh
        echo
        break;;
        [Nn]* ) 
        echo
        read -p "File not created. Import statements should be added to a file with the path examples/resource_directory/import.sh" ; exit;;
        * ) echo; echo "Please answer yes or no.";;
    esac
done
read -p "If you require more than one import statement file go to \"Main Menu\" and then \"Custom\"."
