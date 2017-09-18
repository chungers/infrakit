#!/bin/bash

if [[ $(which terraform | wc -c) -ne 0 ]]
then
	echo "'terraform' has already been installed"
else
	cd ~

	# Prerequisites
	if [ "$(uname)" == "Darwin" ]; then
		brew install jq coreutils
		# For Linux
	elif [ "$(expr substr $(uname -s) 1 5)" == "Linux" ]; then
		sudo apt-get install --assume-yes jq
	fi

	# Get URLs for most recent versions
	# For OS-X
	if [ "$(uname)" == "Darwin" ]; then
		terraform_url=$(curl -s https://releases.hashicorp.com/index.json | jq '{terraform}' | egrep "darwin.*64" | gsort --version-sort -r | head -1 | awk -F[\"] '{print $4}')
		# For Linux
	elif [ "$(expr substr $(uname -s) 1 5)" == "Linux" ]; then
		terraform_url=$(curl -s https://releases.hashicorp.com/index.json | jq '{terraform}' | egrep "linux.*amd64" | sort --version-sort -r | head -1 | awk -F[\"] '{print $4}')
	fi

	# Create a move into directory.
	cd
	mkdir terraform && cd $_

	# Download Terraform. URI: https://www.terraform.io/downloads.html
	echo "Downloading $terraform_url."
	curl -s -o terraform.zip $terraform_url
	# Unzip and install
	unzip -qq terraform.zip

	if [ "$(uname)" == "Darwin" ]; then
		echo '
		# Terraform Path.
		export PATH=~/terraform/:$PATH
		' >>~/.bash_profile

		source ~/.bash_profile
		# For Linux
	elif [ "$(expr substr $(uname -s) 1 5)" == "Linux" ]; then
		echo '
		# Terraform Path
		export PATH=~/terraform/:$PATH
		' >>~/.bashrc

		source ~/.bashrc
	fi
fi
