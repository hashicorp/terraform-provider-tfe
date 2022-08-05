#!/bin/bash

echo
read -p "Add code snippets to demonstrate how your resource is defined in a terraform config file."
while true; do
    echo
    read -p "   -> Generate an empty example file? [y|n] " yn
    case $yn in
        [Yy]* ) 
        # TODO: script to create ex file
        ./createexfile.sh
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
