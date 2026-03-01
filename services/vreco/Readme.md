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

## Fly.io secrets (Blockening account)

To deploy the Blockening account management feature, set `TALO_ACCESS_KEY` and `SESSION_SECRET` as Fly secrets. From the vreco service folder:

```bash
cd services/vreco   # if not already there
go run ./scripts/set-fly-secrets
```

This reads `.env` (handles special characters) and runs `fly secrets set`. Ensure `.env` contains `TALO_ACCESS_KEY` and `SESSION_SECRET` (32+ characters).
