#!/bin/bash

while true; do
    echo
    read -p "   -> Add another example file? [y|n] " yn
    case $yn in
        [Yy]* ) 
        read -p "   -> What is your example filename?" filename
        echo $filename
        # TODO: script to create example file (pass in $filename) and add path to template file
        echo
        read -p "File created at examples/resource_directory/example.tf"
        read -p "Path added to custom template file under # Example Usage"
        ./addexample.sh;;

        [Nn]* ) 
        echo
        read -p "File not created. Examples should be added to a file with the path examples/resource_directory/example.tf" ; break;;
        * ) echo "Please answer yes or no.";;
    esac
done