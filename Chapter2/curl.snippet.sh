curl --location --request POST '127.0.0.1:8080/task/' \
--header 'Content-Type: application/json' \
--data-raw '{
    "id": "8b171ce0-6f7b-4c22-aa6f-8b110c19f83a",
    "name": "A task",
    "description": "A task that need to be executed at the timestamp specified",
    "timestamp": 1645275972000
}'

curl --location --request GET 'localhost:8080/task'

curl --location --request GET 'localhost:8080/task/8b171ce0-6f7b-4c22-aa6f-8b110c19f83a'

