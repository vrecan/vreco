grpcurl -plaintext -d @ localhost:2020 chat.v1.ChatService/SendMessage <<EOM
{
  "msg": "Ben"
}          
EOM


grpcurl -plaintext -d @ localhost:2020 chat.v1.ChatService/GetMessages <<EOM
{
  "limit": 10,
  "start": 0
}
EOM
