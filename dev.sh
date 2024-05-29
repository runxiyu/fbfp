#!/bin/sh
while true
do
	./venv/bin/python3 -m flask --app 'fbfp:make_debug_app()' run --debug
done
