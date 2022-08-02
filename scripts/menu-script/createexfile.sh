#!/bin/bash

read -p "Is this a resource or data source? [r|d]" rd
echo
read -p "Enter your resource name (formatted tfe_resourcename)" resourcename
while true; do
    case $rd in
            [Rr]* ) 
            dir1="resources"; break;;

            [Dd* ) 
            dir1="data-sources";;
            * ) echo "Please answer resource or data source [r|d]."; exit;;
    esac
done

path=../../examples/$dir1/$resourcename
#echo $path
mkdir -p $path
filepath=../../examples/$dir1/$resourcename/resource.tf
#echo $filepath
touch $filepath

# while read dirname others; do
#     mkdir "$dirname"
# done < list.txt

# find . -type d -exec touch {}/import.sh \;
# find . -type d -exec touch {}/resource.tf \;s