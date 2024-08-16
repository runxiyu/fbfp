#!/bin/sh

set -x

signal_handler () {
	echo
	# sudo systemctl stop mariadb
	exit
}

trap signal_handler INT

sudo systemctl start mariadb

while true
do
	./venv/bin/python3 -m flask --app 'fbfp:make_debug_app()' run --debug
done
