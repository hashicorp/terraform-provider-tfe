#!/bin/bash

while true; do
    echo
    case $yn in
        [Yy]* ) 
        read -p "-> What is your example filename?" filename
        echo $filename
        echo
        read -p "File created at examples/resource_directory/example.tf"
        echo
        ./14-cust-examplesinfo.sh; break;;
x
        [Nn]* ) 
        echo
        read -p "File not created. Examples should be added to a file with the path examples/resource_directory/example.tf" ; break;;
        * ) echo "Please answer yes or no.";;
    esac
done

read -p "-> Please enter your resource or datasource name " resourcename
read -p "-> Please enter your example filename " filename
path=../../examples/$dir1/tfe_$resourcename/$filename.tf
#echo $path
mkdir -p $path
filepath=../../examples/$dir1/$filename.tf
printpath=examples/$dir1/tfe_$resourcename/$filename.tf
#echo $filepath
touch $filepath
echo 
read -p "Example file created at: ${printpath}" 