#!/usr/bin/env bash

### Colors ##
ESC=$(printf '\033') RESET="${ESC}[0m" BLACK="${ESC}[30m" RED="${ESC}[31m"
GREEN="${ESC}[32m" YELLOW="${ESC}[33m" BLUE="${ESC}[34m" MAGENTA="${ESC}[35m"
CYAN="${ESC}[36m" WHITE="${ESC}[37m" DEFAULT="${ESC}[39m"

### Color Functions ##

greenprint() { printf "${GREEN}%s${RESET}\n" "$1"; }
blueprint() { printf "${BLUE}%s${RESET}\n" "$1"; }
redprint() { printf "${RED}%s${RESET}\n" "$1"; }
yellowprint() { printf "${YELLOW}%s${RESET}\n" "$1"; }
magentaprint() { printf "${MAGENTA}%s${RESET}\n" "$1"; }
cyanprint() { printf "${CYAN}%s${RESET}\n" "$1"; }
fn_bye() { echo "Bye bye."; exit 0; }
fn_fail() { echo "Wrong option." exit 1; }

echo
read -p "Custom templates are used to add more sections to the basic template..."
read -p "
    NOTE: All markdown changes should be made to .tmpl files in the templates directory. 
        ** Do not directly edit the .md files in the docs directory."
echo
read -p "If you are unsure, first run through Main Menu > Basic."

while true; do
    echo
    read -p "   -> Generate a custom template file? [y|n] " yn
    case $yn in
        [Yy]* ) 
        # TODO: script to create template file
        read -p "
        File created at templates/resources/resource_name.md.tmpl (edit with markdown)"; break;;
        [Nn]* ) 
        echo
        read -p "File not created. Custom templates should be added within the templates directory" ; exit;;
        * ) echo; echo "Please answer yes or no.";;
    esac
done

mainmenu() {
    echo -ne "
$(magentaprint 'COMMON CASES')
$(greenprint '1)') >1 Example File
$(greenprint '2)') >1 Import Statement File
$(greenprint '3)') Notes defined outside of schema descriptions
$(redprint '0)') Exit
Choose an option:  "
    read -r ans
    case $ans in
    1)
        ./addexample.sh
        mainmenu
        ;;
    2)
        ./addimport.sh
        mainmenu
        ;;
    3)
        ./addnote.sh
        mainmenu
        ;;
    0)
        fn_bye
        ;;
    *)
        fn_fail
        ;;
    esac
}

mainmenu