# Vreco
<br>
Main repository behind https://vreco.fly.dev.


# Instructions

## Linux
```bash
sudo apt-get install npm && npm install && \
GO111MODULE=on go install \
  github.com/cosmtrek/air@v1.29.0
~/go/bin/air
```

add to your ~/.bashrc file
```bash
export PATH=$PATH:~/go/bin
```

## OSX
```bash
brew install npm && npm install 
go install github.com/cosmtrek/air@v1.29.0
~/go/bin/air
```
