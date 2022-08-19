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
$(greenprint '1)') ReadMe
$(magentaprint '2)') Return to Main Menu
$(redprint '0)') Exit
Choose an option:  "
    read -r ans
    case $ans in
    1)

        ./12-custominfo.sh
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
$(greenprint '1)') ReadMe
$(greenprint '2)') Return to Basic Menu
$(cyanprint '3)') Return to Main Menu
$(redprint '0)') Exit
Choose an option:  "
    read -r ans
    case $ans in
    1)
        ./11-noteinfo.sh
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
$(greenprint '1)') ReadMe
$(greenprint '2)') Continue to NOTES
$(cyanprint '3)') Go to Basic Menu
$(magentaprint '4)') Go to Main Menu
$(redprint '0)') Exit
Choose an option:  "
    read -r ans
    case $ans in
    1)
        ./8-importinfo.sh
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
$(greenprint '1)') ReadMe
$(greenprint '2)') Continue to IMPORT STATEMENT
$(cyanprint '3)') Go to Basic Menu
$(magentaprint '4)') Go to Main Menu
$(redprint '0)') Exit
Choose an option:  "
    read -r ans
    case $ans in
    1)
        ./5-exampleinfo.sh
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
$(greenprint '1)') ReadMe
$(greenprint '2)') Continue to EXAMPLES
$(cyanprint '3)') Return to Basic Menu
$(magentaprint '4)') Go to Main Menu
$(redprint '0)') Exit
Choose an option:  "
    read -r ans
    case $ans in
    1)
        ./3-descinfo.sh
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
$(greenprint '1)') Descriptions
$(greenprint '2)') Examples
$(greenprint '3)') Import Statement
$(greenprint '4)') Notes
$(greenprint '5)') FINAL STEP (run tfplugindocs)
$(magentaprint '6)') Return to Main Menu
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
        ./run-tfplugindocs.sh
        ;;
    6)
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
$(greenprint '1)') ReadMe
$(greenprint '2)') BASIC
$(greenprint '3)') CUSTOM
$(greenprint '4)') FINAL STEP (run tfplugindocs)
$(redprint '0)') Exit
Choose an option:  "
    read -r ans
    case $ans in
    1)
        ./2-main-menu-info.sh
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
    4) 
        ./run-tfplugindocs.sh
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

