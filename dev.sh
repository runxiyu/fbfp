#!/bin/sh
while true
do
	flask --app 'fbfp:make_debug_app()' run --debug
done
