# Comment utiliser

1. Recupere tes identifiants sur https://developer.spotify.com/dashboard

2. Copie .env.example vers .env et met tes vrais identifiants

3. Charge les variables:
```
source .env
```
ou
```
export SPOTIFY_CLIENT_ID="ton_vrai_id"
export SPOTIFY_CLIENT_SECRET="ton_vrai_secret"
```

4. Lance:
```
go run main.go
```
