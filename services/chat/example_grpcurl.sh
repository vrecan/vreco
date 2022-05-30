grpcurl -plaintext -d @ localhost:2020 chat.v1.ChatService/SayHello <<EOM
{
  "name": "Ben"
}          
EOM
{
  "message": "Hello there Ben"
}
