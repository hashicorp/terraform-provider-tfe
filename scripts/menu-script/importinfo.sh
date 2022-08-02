#!/bin/bash
echo
read -p "Add an example import statement if applicable."
while true; do
    echo
    read -p "   -> Generate an empty import statement file? [y|n] " yn
    case $yn in
        [Yy]* ) 
        # TODO: script to create import file
        echo
        read -p "File created at examples/resource_directory/import.sh"; break;;
        [Nn]* ) 
        echo
        read -p "File not created. Import statements should be added to a file with the path examples/resource_directory/import.sh" ; exit;;
        * ) echo "Please answer yes or no.";;
    esac
done
echo
read -p "If you require more import statement files, go to Main Menu > Custom."
