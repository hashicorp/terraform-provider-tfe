#!/bin/bash
while true; do
    echo
    read -p "-> Is this a resource or data source? [r|d] " rd
    case $rd in
            [Rr]* ) 
            dir1="resources"; break;;

            [Dd]* ) 
            dir1="data-sources"; break;;
            * ) echo; echo "Please answer resource or data source [r|d]. ";; 
    esac
done
echo
read -p "Please enter your resource or datasource name " resourcename
path=../../examples/$dir1/tfe_$resourcename
#echo $path
mkdir -p $path
filepath=../../examples/$dir1/tfe_$resourcename/resource.tf
printpath=examples/$dir1/tfe_$resourcename/resource.tf
#echo $filepath
touch $filepath
echo 
read -p "Example file created at: ${printpath}" 