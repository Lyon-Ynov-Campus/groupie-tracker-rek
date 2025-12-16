# Groupie Tracker - Jeu Multijoueur

Salut ! C'est mon projet de jeux en ligne. J'ai fait ça pour apprendre Go et faire jouer mes potes ensemble.

## C'est quoi ?

C'est un site web où tu peux jouer à des jeux avec tes amis :
- Un blindtest (devine la musique)
- Un petit bac (pas encore fini lol)

Tu crées une salle, tu donnes le code à tes potes et vous jouez ensemble !

## Comment ça marche ?

### Ce qu'il faut avoir sur ton PC

Tu dois installer Go. Va sur https://go.dev/dl/ et télécharge la version pour ton système.

Pour vérifier que c'est bon :
```bash
go version
```

Si ça affiche un truc comme "go version 1.21" c'est bon.

### Installation

1. Télécharge le projet (ou clone le si tu connais git)
2. Ouvre un terminal dans le dossier
3. Lance cette commande :

```bash
go mod download
```

Si ça marche pas, essaye :
```bash
go mod init rek
go mod tidy
```

### Lancer le serveur

Dans le terminal, tape juste :

```bash
go run main.go
```

Tu devrais voir :
```
Base de données initialisée avec succès.
```

Après ouvre ton navigateur et va sur : `http://localhost:8080`

## Comment jouer

### Créer un compte

1. Clique sur "S'inscrire"
2. Mets un nom, un email et un mot de passe
3. Clique sur créer

### Se connecter

1. Clique sur "Se connecter"
2. Tape ton email et mot de passe
3. Tu arrives sur le tableau de bord

### Créer une partie

1. Clique sur "Créer une salle"
2. Choisis le jeu (pour l'instant y'a que le blindtest qui marche bien)
3. Note le code qui s'affiche (genre ABC123)
4. Donne ce code à tes amis

### Rejoindre une partie

1. Ton pote te donne un code
2. Clique sur "Rejoindre une salle"
3. Entre le code
4. Attend que le créateur lance la partie

### Jouer au Blindtest

1. La musique démarre automatiquement
2. Tu as 30 secondes pour deviner le titre
3. Tape juste le titre (pas l'artiste)
4. Clique sur valider
5. À la fin tu vois les scores

## Organisation des fichiers

```
groupie-tracker-rek/
├── main.go              # Le fichier principal qui lance tout
├── rek.db               # La base de données (se crée tout seul)
├── src/                 # Tout le code du serveur
├── templates/           # Les pages HTML
└── static/              # Les CSS et JavaScript
```

### Les fichiers importants

- `main.go` : C'est là que tout commence
- `src/handlers.go` : Gère les pages (accueil, connexion, etc)
- `src/createroom.go` : Pour créer et rejoindre les salles
- `src/blindtest_match.go` : La logique du blindtest
- `src/ws_handler.go` : Les websockets (pour le temps réel)
- `templates/game.html` : La page du jeu
- `static/match_script.js` : Le JavaScript du blindtest

## Les routes (URLs)

### Pages publiques
- `/` : Page d'accueil
- `/connexion` : Se connecter
- `/register` : S'inscrire

### Pages privées (faut être connecté)
- `/dashboard` : Tableau de bord
- `/salle-initialisation` : Créer une salle
- `/rejoindre-salle` : Rejoindre une salle
- `/salle/{code}` : La salle d'attente
- `/game/{code}` : Le jeu

## Technologies

J'ai utilisé :
- Go pour le backend
- SQLite pour la base de données (c'est simple)
- WebSocket pour le temps réel
- HTML/CSS/JavaScript basique pour le front

## Si ça marche pas

### Erreur "cannot find package"
```bash
go mod download
go mod tidy
```

### Le port 8080 est déjà utilisé
Ouvre `main.go` et change la dernière ligne :
```go
http.ListenAndServe(":3000", nil) // Change le 8080 en 3000
```

### La base de données est buguée
Supprime le fichier `rek.db` et relance :
```bash
rm rek.db
go run main.go
```

### Le WebSocket se déconnecte
Rafraîchis la page (F5). Sinon regarde dans la console du navigateur (F12) pour voir l'erreur.

### Pas de son dans le blindtest
- Vérifie ta connexion internet
- Autorise le son dans ton navigateur
- Des fois l'API Deezer bug, relance la partie

## Trucs qui marchent pas encore

- Le bouton "Quitter" dans le jeu marche pas (j'ai oublié de faire la route)
- Le petit bac est pas terminé
- Des fois les routes bugent entre elles (je sais pas trop pourquoi)
- La base de données est pas bien partagée entre les fonctions (je corrigerai)

## Ce que je veux ajouter

- Finir le petit bac
- Corriger le bouton quitter
- Faire un chat dans les salles
- Ajouter des avatars
- Faire un classement général
- Rendre ça plus joli sur mobile

## Bugs connus

1. Si tu quittes la page pendant le jeu, ça bug un peu
2. Parfois les scores s'affichent en double (j'ai pas compris pourquoi)
3. Le timer peut décaler entre les joueurs
4. Si tu rafraîchis pendant une partie, t'es éjecté

## Notes

- Les mots de passe sont chiffrés (bcrypt)
- J'ai fait ça en quelques semaines pour apprendre
- C'est pas parfait mais ça marche plutôt bien
- N'hésite pas à me dire si tu trouves des bugs

## Aide

Si t'as un problème :
1. Regarde les erreurs dans le terminal
2. Ouvre la console du navigateur (F12)
3. Essaye de redémarrer le serveur
4. Vérifie que t'es bien connecté à internet

---

Fait avec ❤️ pour apprendre Go

PS : C'est mon premier gros projet en Go alors soyez indulgents ! 😅