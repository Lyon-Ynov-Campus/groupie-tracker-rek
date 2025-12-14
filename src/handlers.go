package server

// Dans AfficherSalleHandler, ajoute le dispatch /salle/{code}/start juste après celui de /config :

// /salle/{code}/config
if len(parts) >= 2 && parts[1] == "config" {
    ConfigurerSalleHandler(w, r, code)
    return
}

// /salle/{code}/start
if len(parts) >= 2 && parts[1] == "start" {
    StartBlindtestHandler(w, r, code)
    return
}