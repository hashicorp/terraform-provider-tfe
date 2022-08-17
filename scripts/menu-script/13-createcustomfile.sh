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
read -p "Please enter your resource or datasource name. " resourcename
path=../../templates/$dir1
#echo $path
mkdir -p $path
filepath=../../templates/$dir1/$resourcename.tmpl
printpath=templates/$dir1/$resourcename.tmpl
#echo $filepath
touch $filepath 
cp template.txt $filepath
read -p "Template file created at: ${printpath}" 

