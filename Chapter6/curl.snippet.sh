#!/bin/bash

while true; do
    curl --location --request POST 'localhost:8080/task/' \
	 --header 'Content-Type: application/json' \
	 --data-raw '{
	 	    "id": "'$(date +%s%N)'",
		    "name": "A task",
		    "description": "A task needed to be executed at the timestamp specified",
		    "timestamp": 1645275972000,
		    "location": {
		    		"id": "8b171ce0-6f7b-4c22-aa6f-8b110c19f83a",
				"name": "Liverpoll Street Station",
				"description": "Station for Tube and National Rail",
				"longitude": -0.81966,
				"latitude": 51.517336
				}
		    }'
sleep 1
done
