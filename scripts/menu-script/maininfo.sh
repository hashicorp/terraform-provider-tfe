#!/bin/bash
echo
read -p "Basic vs. Custom Template"
echo
read -p "Basic templates follow a predefined format: 

    # resource_name
        - resource description 
        - note (opt)

    ## Example Usage
        - 1 example code snippet only

    ## Schema
        - resource attributes and descriptions
        - notes (opt)

    ## Import (opt)
        - 1 import statement only 

"
echo 
read -p "Custom templates should be used to add any additional headings or multiple example code snippets."
read -p "If you are unsure, please go back to main menu and first select Basic." 

