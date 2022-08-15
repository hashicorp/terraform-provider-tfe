#!/bin/bash
echo
echo
read -p "Add code snippets to demonstrate how your resource is declared in a terraform config file with blocks, arguments, and expressions."
echo
read -p "Here is an example from examples/resource_tfe_agent_pool/resource.tf: " 
while true; do
    echo
    read -p "   -> Generate an empty example file? [y|n] " yn
    case $yn in
        [Yy]* ) 
        ./6-createexfile.sh
        echo
        read -p "File created."; break;;
        [Nn]* ) 
        echo
        read -p "File not created. Examples should be added to a file with the path examples/resource_directory/example.tf" ; exit;;
        * ) echo "Please answer yes or no.";;
    esac
done
echo
read -p "If you require more examples or headings, go to Main Menu > Custom."
