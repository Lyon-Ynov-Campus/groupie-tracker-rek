# 🎮 Groupie Tracker - Application de Jeux Multijoueurs

## 📋 Table des matières
- [Vue d'ensemble](#vue-densemble)
- [Fonctionnalités](#fonctionnalités)
- [Architecture du projet](#architecture-du-projet)
- [Prérequis](#prérequis)
- [Installation](#installation)
- [Lancement de l'application](#lancement-de-lapplication)
- [Utilisation](#utilisation)
- [Structure du code](#structure-du-code)
- [Routes API](#routes-api)
- [Technologies utilisées](#technologies-utilisées)
- [Dépannage](#dépannage)

---

## 🎯 Vue d'ensemble

**Groupie Tracker** est une application web multijoueurs en temps réel permettant de jouer à différents jeux :
- 🎵 **Blindtest musical** (devinez les chansons)
- 📝 **Petit Bac** (trouvez des mots par catégorie)

L'application permet aux utilisateurs de créer des salles de jeu, d'inviter des amis via un code, et de jouer ensemble en temps réel grâce aux WebSockets.

---

## ✨ Fonctionnalités

### 🔐 Authentification
- ✅ Inscription avec email et mot de passe
- ✅ Connexion sécurisée
- ✅ Sessions persistantes avec cookies
- ✅ Déconnexion

### 🎮 Gestion des salles
- ✅ Création de salle privée avec code unique
- ✅ Rejoindre une salle via code
- ✅ Configuration de la salle (type de jeu, paramètres)
- ✅ Liste des joueurs en temps réel
- ✅ Démarrage de partie par le créateur

### 🎵 Blindtest Musical
- ✅ Écoute d'extraits musicaux
- ✅ Devinez le titre de la chanson
- ✅ Système de points
- ✅ Timer pour chaque manche
- ✅ Classement en fin de partie

### 📝 Petit Bac (prévu)
- 🔄 Jeu de mots par catégorie
- 🔄 Rounds chronométrés
- 🔄 Système de validation

---

## 🏗️ Architecture du projet

```
groupie-tracker-rek/
├── main.go                 # Point d'entrée, configuration des routes
├── rek.db                  # Base de données SQLite (créée automatiquement)
│
├── src/                    # Code source backend (package server)
│   ├── handlers.go         # Handlers HTTP principaux (accueil, auth)
│   ├── database.go         # Initialisation et gestion de la DB
│   ├── security.go         # Middleware d'authentification
│   ├── user.go             # Logique utilisateur
│   ├── createroom.go       # Création de salles
│   ├── http_game.go        # Handlers de jeu
│   ├── http_api.go         # API REST pour les salles
│   ├── ws_handler.go       # Gestion des WebSockets
│   ├── ws_hub.go           # Hub de connexions WebSocket
│   ├── ws_types.go         # Types pour WebSocket
│   ├── blindtest_match.go  # Logique du blindtest
│   ├── blindtest_runtime.go # Exécution du blindtest
│   ├── blindtest_deezer_genre.go # Intégration API Deezer
│   ├── petitbac_logic.go   # Logique du Petit Bac
│   ├── gameconfig.go       # Configuration des jeux
│   ├── membre.go           # Gestion des membres de salle
│   ├── score.go            # Calcul et gestion des scores
│   ├── query.go            # Requêtes SQL
│   └── render.go           # Rendu des templates HTML
│
├── templates/              # Templates HTML
│   ├── accueil.html        # Page d'accueil
│   ├── authentification.html # Page de connexion/inscription
│   ├── landingpage.html    # Dashboard après connexion
│   ├── init_room.html      # Création de salle
│   ├── salle.html          # Salle d'attente
│   ├── config_salle.html   # Configuration de salle
│   ├── game.html           # Interface blindtest
│   └── petitbac.html       # Interface petit bac
│
└── static/                 # Fichiers statiques
    ├── styles.css          # CSS global
    ├── init_salle.css      # CSS pour les salles/jeux
    ├── landingpage.css     # CSS du dashboard
    ├── ws_room.js          # WebSocket salle d'attente
    ├── match_script.js     # Logique blindtest client
    └── match_petitbac.js   # Logique petit bac client
```

---

## 🔧 Prérequis

Avant de commencer, assurez-vous d'avoir installé :

### 1. **Go (Golang)**
- Version minimale : **Go 1.19+**
- Téléchargement : https://go.dev/dl/

Vérifiez l'installation :
```bash
go version
```

### 2. **Git** (optionnel, pour cloner le projet)
```bash
git --version
```

### 3. **Un navigateur web moderne**
- Chrome, Firefox, Edge, Safari (version récente)

---

## 📥 Installation

### Étape 1 : Cloner ou télécharger le projet

**Option A : Avec Git**
```bash
git clone <url-du-repo>
cd groupie-tracker-rek
```

**Option B : Sans Git**
1. Téléchargez le ZIP du projet
2. Extrayez-le dans un dossier
3. Ouvrez un terminal dans ce dossier

### Étape 2 : Installer les dépendances Go

```bash
go mod download
```

Si le fichier `go.mod` n'existe pas, créez-le :
```bash
go mod init rek
go mod tidy
```

### Étape 3 : Vérifier les dépendances requises

Le projet utilise :
- `github.com/gorilla/websocket` (WebSockets)
- `github.com/mattn/go-sqlite3` (base de données)
- `golang.org/x/crypto/bcrypt` (hashage des mots de passe)

Ces dépendances s'installent automatiquement avec `go mod download`.

---

## 🚀 Lancement de l'application

### Démarrage du serveur

```bash
go run main.go
```

Vous devriez voir :
```
Base de données initialisée avec succès.
Serveur démarré sur :8080
```

### Accéder à l'application

Ouvrez votre navigateur et allez sur :
```
http://localhost:8080
```

---

## 📖 Utilisation

### 1️⃣ **Créer un compte**

1. Sur la page d'accueil, cliquez sur **"S'inscrire"**
2. Remplissez le formulaire :
   - Nom d'utilisateur
   - Email
   - Mot de passe
3. Cliquez sur **"Créer un compte"**

### 2️⃣ **Se connecter**

1. Cliquez sur **"Se connecter"**
2. Entrez vos identifiants
3. Vous êtes redirigé vers le **Dashboard**

### 3️⃣ **Créer une salle de jeu**

1. Sur le dashboard, cliquez sur **"Créer une salle"**
2. Choisissez le type de jeu :
   - 🎵 Blindtest
   - 📝 Petit Bac
3. Configurez les paramètres (nombre de manches, durée, etc.)
4. Cliquez sur **"Créer"**
5. **Notez le code de la salle** (ex: `ABC123`)

### 4️⃣ **Rejoindre une salle**

1. Sur le dashboard, cliquez sur **"Rejoindre une salle"**
2. Entrez le **code de la salle** reçu
3. Cliquez sur **"Rejoindre"**

### 5️⃣ **Jouer au Blindtest**

1. Dans la salle d'attente, attendez que le créateur clique sur **"Démarrer la partie"**
2. Une fois le jeu lancé :
   - 🎵 Écoutez l'extrait musical
   - ⏱️ Vous avez 30 secondes par manche
   - ✍️ Tapez le **titre de la chanson** dans le champ
   - ✅ Validez votre réponse
3. À la fin, consultez le **classement final**

### 6️⃣ **Quitter/Rejouer**

- **Rejouer** : Cliquez sur "Rejouer" pour une nouvelle partie
- **Quitter** : Cliquez sur "Quitter la salle" ou "Retour salle"
- **Déconnexion** : Cliquez sur "Se déconnecter" dans le dashboard

---

## 🔍 Structure du code

### Backend (Go)

#### **main.go**
Point d'entrée qui :
- Initialise la base de données SQLite
- Configure toutes les routes HTTP
- Démarre le serveur sur le port `:8080`

#### **src/handlers.go**
Handlers principaux :
- `HomeHandler` : Page d'accueil
- `RegisterHandler` : Inscription
- `LoginHandler` : Connexion
- `LogoutHandler` : Déconnexion
- `LandingPageHandler` : Dashboard

#### **src/security.go**
- `RequireAuth()` : Middleware vérifiant l'authentification
- Gestion des sessions via cookies

#### **src/createroom.go**
- `CreerSalleHandler` : Crée une salle avec code unique
- `RejoindreSalleHandler` : Rejoint une salle existante
- `AfficherSalleHandler` : Affiche la salle d'attente

#### **src/ws_handler.go & ws_hub.go**
- Gestion des WebSockets en temps réel
- Broadcast des messages aux joueurs connectés
- Synchronisation de l'état de la salle

#### **src/blindtest_*.go**
- `blindtest_match.go` : Structure d'une partie
- `blindtest_runtime.go` : Exécution du jeu
- `blindtest_deezer_genre.go` : Récupération de musiques via API Deezer

#### **src/database.go**
Tables SQL :
```sql
- users (id, username, email, password_hash)
- sessions (token, user_id, expires_at)
- salles (code, creator_id, game_type, config)
- membres (salle_code, user_id, score)
- blindtest_matches (id, salle_code, state)
```

### Frontend (HTML/JS/CSS)

#### **Templates HTML**
- Utilisation de `html/template` de Go
- Variables injectées : `{{.Username}}`, `{{.Code}}`, etc.

#### **JavaScript**
- `ws_room.js` : WebSocket pour la salle d'attente
- `match_script.js` : Logique du blindtest côté client
- `match_petitbac.js` : Logique du petit bac

---

## 🛣️ Routes API

### Routes publiques
```
GET  /                    → Page d'accueil
GET  /connexion           → Page de connexion
POST /login               → Traitement connexion
GET  /register            → Page d'inscription
POST /register            → Traitement inscription
GET  /static/*            → Fichiers CSS/JS
```

### Routes authentifiées (nécessitent connexion)
```
GET  /dashboard           → Dashboard utilisateur
GET  /logout              → Déconnexion

GET  /salle-initialisation → Formulaire création salle
POST /creer-salle         → Créer une salle
POST /rejoindre-salle     → Rejoindre une salle

GET  /salle/{code}        → Salle d'attente
POST /salle/{code}/start  → Démarrer la partie
POST /salle/{code}/leave  → Quitter la salle

GET  /game/{code}         → Interface de jeu
GET  /api/salle/{code}    → API REST infos salle
WS   /ws/salle/{code}     → WebSocket salle
```

---

## 🛠️ Technologies utilisées

### Backend
- **Go 1.19+** : Langage serveur
- **net/http** : Serveur HTTP natif
- **SQLite3** : Base de données embarquée
- **gorilla/websocket** : WebSockets en temps réel
- **bcrypt** : Hashage sécurisé des mots de passe

### Frontend
- **HTML5** : Structure des pages
- **CSS3** : Mise en forme
- **JavaScript ES6+** : Logique client
- **WebSocket API** : Communication temps réel

### APIs externes
- **Deezer API** : Récupération de musiques pour le blindtest

---

## 🐛 Dépannage

### ❌ Erreur : `cannot find package`
```bash
go mod download
go mod tidy
```

### ❌ Port 8080 déjà utilisé
Modifiez dans `main.go` :
```go
http.ListenAndServe(":3000", nil) // Changez le port
```

### ❌ Base de données verrouillée
```bash
rm rek.db
go run main.go  # Recrée la DB
```

### ❌ WebSocket déconnecté
- Vérifiez que JavaScript est activé
- Vérifiez la console du navigateur (F12)
- Rafraîchissez la page (F5)

### ❌ "Échec d'authentification"
- Supprimez les cookies du site
- Reconnectez-vous

### ❌ Musique ne joue pas (Blindtest)
- Vérifiez votre connexion internet (API Deezer)
- Autorisez le son dans votre navigateur
- Vérifiez que l'API Deezer est accessible

---

## 📝 Notes importantes

### ⚠️ Problèmes connus

1. **Route `/salle/{code}/leave` manquante** dans `main.go`
   - Le bouton "Quitter" dans `game.html` ne fonctionne pas actuellement
   - **Solution temporaire** : Utilisez "Retour salle" ou fermez l'onglet

2. **Base de données non partagée entre handlers**
   - La variable `db` n'est pas accessible dans les handlers
   - Fonctionnera uniquement si `InitDB()` stocke `db` globalement dans le package `server`

3. **Ordre des routes**
   - Les routes `/api/salle/` et `/ws/salle/` peuvent ne jamais être atteintes
   - `/salle/` capture toutes les requêtes commençant par `/salle/`

### 🔒 Sécurité

- ✅ Mots de passe hashés avec bcrypt
- ✅ Sessions sécurisées avec tokens
- ⚠️ Pas de HTTPS (à activer en production)
- ⚠️ Pas de rate limiting (à implémenter)

### 🚀 Améliorations futures

- [ ] Corriger la route `/salle/{code}/leave`
- [ ] Implémenter le jeu Petit Bac complet
- [ ] Ajouter des avatars utilisateurs
- [ ] Historique des parties jouées
- [ ] Classement global
- [ ] Chat dans les salles
- [ ] Support mobile optimisé

---

## 👥 Contribution

Pour contribuer au projet :
1. Forkez le repository
2. Créez une branche (`git checkout -b feature/amelioration`)
3. Committez vos changements (`git commit -m 'Ajout fonctionnalité'`)
4. Pushez (`git push origin feature/amelioration`)
5. Ouvrez une Pull Request

---

## 📄 Licence

Ce projet est un projet éducatif. Tous droits réservés.

---

## 📞 Support

En cas de problème :
1. Vérifiez la section [Dépannage](#dépannage)
2. Consultez les logs du serveur dans le terminal
3. Vérifiez la console du navigateur (F12 → Console)
4. Contactez l'équipe de développement

---

**Bon jeu ! 🎮🎵**