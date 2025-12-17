
function renderScoreboard() {
  scoreList.innerHTML = "";
  if (!state.players) return;

  // Pas de tri. On affiche la liste telle qu'elle arrive du serveur.
  // Le JS ne prend aucune décision.
  state.players.forEach((player) => {
    
    // Détection visuelle pour "Moi" (autorisé car c'est de l'UX)
    const isMe = (player.UserID === state.userID);
    
    // Création de l'élément (Format "Carte")
    const li = document.createElement('li');
    li.className = `score-card ${isMe ? 'is-me' : ''}`;
    
    // Couleur avatar basée sur l'ID (Purement esthétique)
    const colorClass = `av-${player.UserID % 5}`;
    const initial = player.Pseudo.charAt(0).toUpperCase();

    li.innerHTML = `
      <div class="avatar-circle ${colorClass}">${initial}</div>
      <div class="score-name">${player.Pseudo}</div>
      <div class="score-value">${player.Score}</div>
      <div class="score-label">POINTS</div>
    `;

    scoreList.appendChild(li);
  });
}