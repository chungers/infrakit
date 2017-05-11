#!/bin/sh

RED="\033[31m"
GREEN="\033[32m"
YELLOW="\033[33m"
NORMAL="\033[0;39m"

for test in /tests/*_*.sh; do
  printf "${YELLOW}[RUN ]${NORMAL} ${test}\n"

  ${test}

  if [ $? -eq 0 ]; then
    printf "\r${GREEN}[PASS]${NORMAL} ${test}\n"
  else
    printf "\r${RED}[FAIL]${NORMAL} ${test}\n"
    exit 1
  fi
done
