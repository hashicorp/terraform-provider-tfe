#!/bin/bash

while true; do
    echo
    read -p "   -> Add another import file? [y|n] " yn
    case $yn in
        [Yy]* ) 
        read -p "-> What is your import filename?" filename
        echo $filename
        # TODO: script to create import file (pass in $filename) and add path to template file
        echo
        read -p "File created."
        read -p "Path added to custom template file under # Import Statement"
        ./addexample.sh;;

        [Nn]* ) 
        echo
        read -p "File not created. Examples should be added to a file with the path examples/resource_directory/import.sh" ; break;;
        * ) echo "Please answer yes or no.";;
    esac
done

