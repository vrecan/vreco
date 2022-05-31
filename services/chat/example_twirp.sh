curl --request "POST" --header "Content-Type: application/json" \
--data '{"msg": "woo"}' \
http://localhost:8080/twirp/chat.v1.ChatService/SendMessage

curl --request "POST" --header "Content-Type: application/json" \
--data '{}' \
http://localhost:8080/twirp/chat.v1.ChatService/GetMessages
