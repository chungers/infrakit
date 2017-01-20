#!/bin/bash

function runTest {
  TEST_NAME=$1

  cd /tests
  ./test_${TEST_NAME}.sh

  if [ $? -eq 0 ]; then
    echo -e "\e[0;32mPASS\e[0m ${TEST_NAME}"
  else
    echo -e "\e[0;31mFAIL\e[0m ${TEST_NAME}"
  fi
}

if [ "all" == "$1" ]; then
  for file in $(ls test_*.sh); do
    test=${file/test_/}
    test=${test/.sh/}

    runTest ${test}
  done
else
  runTest $1
fi
