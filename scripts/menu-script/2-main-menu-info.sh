#!/bin/bash
echo
read -p "BASIC vs. CUSTOM Template"
echo
read -p "-> BASIC templates generate docs with the following predefined format: "
echo 
read -p "1. Resource Name
        - description 
        - additional notes"
echo
read -p " 2. Example Usage
        - one sample code snippet (showing how to declare the resource in a TF config file)"
echo 
read -p " 3. Attributes Summary  
        - attribute descriptions (pulled from the resource schema)
        - additional notes"
echo 
read -p "4. Import
        - one import statement (if applicable)"

echo 
read -p "-> CUSTOM templates are used to add additional headings or multiple example code snippets."
echo
read -p "If you are unsure, go to \"Main Menu\" and then \"BASIC\"."