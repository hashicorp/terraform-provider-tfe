#!/bin/bash
echo 
read -p "Generate tfe-provider documentation for resources and data sources. Press enter to continue..."
echo
read -p "Each resource and data source has its own web page hosted by the Registry. The basic format is:"

read -p " 
    1. General description
    2. Example Usage (sample code snippets of implementation in a .tf config file)
    3. Schema Outline (attributes and their descriptions)
    4. Import Statement (optional)
"

read -p "~~ PART ONE: ~~"
echo
read -p "General and attribute descriptions are included within the resource schemas defined in the tfe directory."
read -p "   -> Please add your descriptions, then press enter to continue..."
read -p "   -> Notes or warnings can also be added to the descriptions (as strings) with the following formatting: "
echo 
read -p "           \n\n ~> **NOTE:** Your note here"
echo
read -p "   -> If you require notes that cannot be included in any descriptions, press z to create a custom template. "
echo
read -p "~~ PART TWO ~~"
echo
read -p "Add code snippets, to the example file, demonstrating how your resource is defined in a terraform config file."
read -n1 -p "Generate an example file? [y/n]" doit 
if [$doit == 'y']
then 
    echo "runs a script to generate empty file"
     echo "Example file has been created, please add your code snippet in examples/resource_directory/example.tf"
echo
read -p "After adding your example, press enter to continue"
read -p "If you require more examples or headings, press z to create a custom template. Otherwise, press enter to continue."
echo
read -p "Part Three: Add an example import statement if applicable" 
echo
read -n1 -p "Generate import file? [y,n]" doit 
case $doit in
    y|Y) echo "this will generate import file"; echo "Import file has been created, please add your code snippet in examples/resource_directory/import.sh" ;;
    n|N) echo "Import file not created. Press enter to continue..."
esac 
echo
read -p "If you require any addtitional sections, press z to create a custom template"





