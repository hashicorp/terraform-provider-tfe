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


custommenu() {
    echo -ne "
$(cyanprint 'CUSTOM MENU')
$(greenprint '1)') Info
$(magentaprint '2)') Return to Main Menu
$(redprint '0)') Exit
Choose an option:  "
    read -r ans
    case $ans in
    1)
        # TODO: add info script
        ./custominfo.sh
        custommenu
        ;;

    2)
        mainmenu
        custommenu
        ;;
    0)
        fn_bye
        ;;
    *)
        fn_fail
        ;;
    esac
}

notemenu() {
    echo -ne "
$(yellowprint 'NOTES MENU')
$(greenprint '1)') Info
$(greenprint '2)') Return to Basic Menu
$(cyanprint '3)') Return to Main Menu
$(redprint '0)') Exit
Choose an option:  "
    read -r ans
    case $ans in
    1)
        # TODO: info script 
        ./noteinfo.sh
        notemenu
        ;;
    2)
        basicmenu
        ;;
    3)
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

importmenu() {
    echo -ne "
$(yellowprint 'IMPORT STATEMENT MENU')
$(greenprint '1)') Info
$(greenprint '2)') Continue to NOTES
$(cyanprint '3)') Return to Basic Menu
$(magentaprint '4)') Return to Main Menu
$(redprint '0)') Exit
Choose an option:  "
    read -r ans
    case $ans in
    1)
        # TODO: info script 
        ./importinfo.sh
        importmenu
        ;;
    2)
        notemenu
        importmenu
        ;;
    3)
        basicmenu
        ;;
    4)
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

examplemenu() {
    echo -ne "
$(yellowprint 'EXAMPLES MENU')
$(greenprint '1)') Info
$(greenprint '2)') Continue to IMPORT STATEMENT
$(cyanprint '3)') Return to Basic Menu
$(magentaprint '4)') Return to Main Menu
$(redprint '0)') Exit
Choose an option:  "
    read -r ans
    case $ans in
    1)
        # TODO: info script 
        ./exampleinfo.sh
        examplemenu
        ;;
    2)
        importmenu
        examplemenu
        ;;
    3)
        basicmenu
        ;;
    4)
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

descmenu() {
    echo -ne "
$(yellowprint 'DESCRIPTIONS MENU')
$(greenprint '1)') Info
$(greenprint '2)') Continue to EXAMPLES
$(cyanprint '3)') Return to Basic Menu
$(magentaprint '4)') Return to Main Menu
$(redprint '0)') Exit
Choose an option:  "
    read -r ans
    case $ans in
    1)
        ./descinfo.sh
        descmenu
        ;;
    2)
        examplemenu
        descmenu
        ;;
    3)
        basicmenu
        ;;
    4)
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

basicmenu() {
    echo -ne "
$(cyanprint 'BASIC MENU')
$(greenprint '1)') DESCRIPTIONS
$(greenprint '2)') EXAMPLES
$(greenprint '3)') IMPORT STATEMENT
$(greenprint '4)') NOTES
$(magentaprint '5)') Return to Main Menu
$(redprint '0)') Exit
Choose an option:  "
    read -r ans
    case $ans in
    1)
        descmenu
        basicmenu
        ;;
   
    2) 
        examplemenu
        basicmenu
        ;;

    3)
        importmenu
        basicmenu
        ;;

    4)
        notemenu
        basicmenu
        ;;

    5)
        mainmenu
        basicmenu
        ;;
    0)
        fn_bye
        ;;
    *)
        fn_fail
        ;;
    esac
}

mainmenu() {
    echo -ne "
$(magentaprint 'MAIN MENU')
$(greenprint '1)') Info
$(greenprint '2)') Basic
$(greenprint '3)') Custom
$(redprint '0)') Exit
Choose an option:  "
    read -r ans
    case $ans in
    1)
        ./maininfo.sh
        mainmenu
        ;;
    2)
        basicmenu
        mainmenu
        ;;
    3)
        custommenu
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

