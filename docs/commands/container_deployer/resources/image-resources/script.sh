#!/bin/sh

# read input parameter
IMPORTS=$(cat $IMPORTS_PATH)

# read sleep time before, only introduced to prolongate the command such that you could inspect the running
# container before the program is executed
sleepTimeBefore=$(echo $IMPORTS | jq .sleepTimeBefore)
sleep $sleepTimeBefore

# if $OPERATION = "RECONCILE" append the input word, otherwise clean up
if [ $OPERATION == "RECONCILE" ]; then
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
else
  # do some cleanup, not really necessary because the state will not be kept anyhow
  rm $STATE_PATH/state.txt
fi

# read sleep time after, only introduced to prolongate the command such that you could inspect the running
# container before the program is executed
sleepTimeAfter=$(echo $IMPORTS | jq .sleepTimeAfter)
sleep $sleepTimeAfter