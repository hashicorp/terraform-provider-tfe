#!/bin/bash
echo
read -p "Please enter your resource or datasource name (tfe_resourcename) " resourcename
path=../../examples/$dir1/$resourcename
#echo $path
mkdir -p $path
filepath=../../examples/$dir1/$resourcename/import.sh
printpath=examples$dir1/$resourcename/import.sh
#echo $filepath
touch $filepath
echo 
read -p "Import statement file created at: ${printpath}" 