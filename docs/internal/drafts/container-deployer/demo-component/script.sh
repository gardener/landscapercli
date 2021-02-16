#!/bin/sh

IMPORTS=$(cat $IMPORTS_PATH)

sleepTimeBefore=$(echo $IMPORTS | jq .sleepTimeBefore)
sleep $sleepTimeBefore

# read import parameter word and remove double quotes
inputWord=$(echo $IMPORTS | jq .word)
inputWord=$(echo $inputWord | sed 's/"//g')

# create empty state file if it does not exists
[[ ! -f $STATE_PATH/state.txt ]] && echo "" > $STATE_PATH/state.txt

# append word to state
state=$(cat $STATE_PATH/state.txt)
state="$state $inputWord"
echo $state > $STATE_PATH/state.txt

# write export values
output="{\"sentence\": \"$state\"}"
echo $output > $EXPORTS_PATH

sleepTimeAfter=$(echo $IMPORTS | jq .sleepTimeAfter)
sleep $sleepTimeAfter