#!/bin/bash

for FILE in *
    do  VAR=$(awk '/shell/{ print NR; exit }' $FILE)
        VARPLUSONE=$(($VAR+1))
        
        echo $VARPLUSONE
        VAR2=$(awk NR==${VAR} $FILE)
     
        
done