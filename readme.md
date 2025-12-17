# 🎵 HabiBeats - REK : Le jeu multi entre potes

Bienvenue sur **HabiBeats** !  
C’est un projet qu’on a codé à plusieurs pour apprendre Go, s’amuser, et faire jouer nos potes ensemble.  
Ici, tu peux lancer un **Blindtest** ou un **Petit Bac** , inviter tes amis, et voir qui est le boss du game.

---

## 🚀 C’est quoi ce projet ?

Un site web où tu peux :

- Créer un compte (**alerte spoiler** : sans compte tu peux pas jouer mon pote)
- Créer ou rejoindre une salle
- Mets un pseudo, un mail, un mot de passe (⚠️ il te faudra respecter le règlement CNIL : ton mot de passe doit faire au moins 8 caractères, contenir une majuscule, une minuscule, un chiffre et un caractère spécial… oui, c’est relou, mais c’est la loi !)

## 🎮 Jouer

### 1. Créer ou rejoindre une salle

- Tu te connectes, tu choisis ton jeu (Blindtest ou Petit Bac)
- Clique sur “Créer une salle” ou “Rejoindre une salle”
- Invite tes amis avec le code de la salle
- Pour le Blindtest, tu choisis le type de musique : rap, pop ou rock (désolé pour les fans de jazz, on fera mieux la prochaine fois 😅)

---

## 🛠️ Comment installer et lancer le projet

### 1. Prérequis

- **Go** (version 1.21 ou plus récent)
- Un navigateur web (Chrome, Firefox, Edge, …)
- (Optionnel) Des potes pour jouer avec toi 😁

### 2. Récupérer le projet

Clone ou télécharge ce repo :

```bash
git clone https://github.com/tonuser/groupie-tracker-rek.git
cd groupie-tracker-rek
```

### 3. Installer les dépendances Go

```bash
go mod download
```

Si ça râle, tente :

```bash
go mod init rek
go mod tidy
```

### 4. Lancer le serveur

```bash
go run main.go
```

Tu dois voir :  
`Base de données initialisée avec succès.`

### 5. Ouvrir le site

Va sur [http://localhost:8080](http://localhost:8080) dans ton navigateur.

---

## 👤 Créer un compte

- Clique sur “S’inscrire”
- Mets un pseudo, un mail, un mot de passe
- Valide, puis connecte-toi

---

## 🎮 Jouer

### 1. Créer ou rejoindre une salle

- Clique sur “Créer une salle” ou “Rejoindre une salle”
- Choisis ton jeu (Blindtest ou Petit Bac)
- Invite tes amis avec le code de la salle

### 2. Blindtest

- Le jeu lance un extrait musical à chaque manche
- Tape ta réponse (titre ou artiste)
- Plus tu réponds vite, plus tu gagnes de points
- Le scoreboard s’affiche à la fin

### 3. Petit Bac

- Une lettre et des catégories s’affichent
- Remplis tes réponses le plus vite possible
- Ensuite, tu votes sur les réponses des autres (valide ou refusé)
- Le scoreboard s’affiche à la fin

- **Bonus** : Quand tu crées une salle Petit Bac, tu peux choisir les catégories (Artiste, Album, Groupe de musique.... ), en ajouter ou en supprimer comme tu veux avant de lancer la partie !
- Si tu enregistres tes catégories, il faudra revenir à la salle pour commencer le jeu (un bouton est prévu pour ça)
- Et si tu t’es trompé de jeu, pas de panique : tu peux toujours revenir au choix du jeu grâce à un bouton "Changer de jeu"

---

## 🖌️ Le design

- CSS moderne, responsive, avec un peu de glow et de fun
- Scoreboard stylé, avatars colorés, tout pour l’ambiance

---

## 🗂️ Structure du projet

```
groupie-tracker-rek/
│
├── main.go                  # Point d’entrée du serveur Go
├── go.mod                   # Dépendances Go
│
├── src/                     # Code backend (Go)
│   ├── createroom.go        # Logique création de salle
│   ├── handlers.go          # Handlers HTTP principaux
│   ├── gameconfig.go        # Config des jeux (catégories, playlists)
│   ├── petitbac_logic.go    # Logique du jeu Petit Bac
│   ├── blindtest_runtime.go # Logique du jeu Blindtest
│   ├── ws_handler.go        # WebSocket handler
│   ├── ...                  # (autres fichiers Go)
│
├── templates/               # Templates HTML (Go)
│   ├── accueil.html         # Page d’inscription
│   ├── authentification.html# Page de connexion
│   ├── landingpage.html     # Sélection du jeu
│   ├── salle.html           # Salle d’attente
│   ├── game.html            # Blindtest
│   ├── petitbac.html        # Petit Bac
│   ├── ...                  # (autres pages)
│
├── static/                  # Fichiers statiques (CSS, JS, images)
│   ├── init_salle.css       # Style principal
│   ├── scoreboard.css       # Style du scoreboard
│   ├── landingpage.css      # Style de la page d’accueil
│   ├── match_script.js      # JS du Blindtest
│   ├── match_petitbac.js    # JS du Petit Bac
│   ├── scoreboard_render.js # Rendu JS du scoreboard (commun)
│   ├── ...                  # (autres assets)
│
└── readme.md                # Ce fichier !
```

- **src/** : tout le backend Go (logique, API, WebSocket, BDD…)
- **templates/** : les pages HTML générées côté serveur
- **static/** : tout ce qui est chargé côté client (CSS, JS, images)
- **main.go** : le serveur web qui lance tout

---

Tu veux comprendre ou modifier un truc ?  
→ Cherche dans le dossier qui correspond à ce que tu veux toucher (backend, frontend, style, etc.)  
→ Et si tu galères, demande à un pote ou ouvre une issue !

---

## 🐞 Bugs connus / TODO

- Le projet est en mode “apprentissage”, donc il peut y avoir des bugs (n’hésite pas à ouvrir une issue ou à corriger !)
- Le code est perfectible, mais il fait le taf pour jouer entre amis

---

## 🙏 Remerciements

Merci à tous ceux qui ont testé, donné des idées, ou juste mis l’ambiance pendant les parties !  
Projet fait avec ❤️ par Ryan, Kerem et Edvige 

---

## 📢 Disclaimer

C’est un projet étudiant, pas une app pro.  
Si tu veux t’en inspirer, go ! Si tu veux l’améliorer, encore mieux !

---

Bon jeu ! 🎉