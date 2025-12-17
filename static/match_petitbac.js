const roomCode = document.body.dataset.roomCode;
// DOM Elements
const statusEl = document.getElementById('status');
const timerEl = document.getElementById('timer');
const letterEl = document.getElementById('letter');
const roundEl = document.getElementById('round');
const answersForm = document.getElementById('answersForm');
const categoriesDiv = document.getElementById('categories');
const votesForm = document.getElementById('votesForm');
const votesDiv = document.getElementById('votes');
const scoreboard = document.getElementById('scoreboard');
const scoreList = document.getElementById('scoreList');

let ws;
let state = null;
let debounceTimer = null; // Pour l'auto-save

// Mémoire locale pour l'interface
let localAnswers = {}; 
let localVotes = {};   

// Variables de suivi pour le rendu
let renderedRound = -1;
let renderedPhase = null;

// --- GESTION ÉTAT ---

function fetchState() {
  fetch(`/api/salle/${roomCode}/petitbac/state`)
    .then(r => r.json())
    .then(updateUI)
    .catch(err => console.warn("Erreur fetch state:", err));
}

function updateUI(newState) {
  if (!newState || !newState.phase) return;
  state = newState;

  updateHeaderInfo();
  handlePhaseDisplay();

  if (state.phase === "playing") {
    renderCategoriesOnce();
  } else if (state.phase === "validation") {
    renderVotesSmart();
  } else if (state.phase === "finished") {
    renderScoreboard();
  }
}

function updateHeaderInfo() {
  if (state.endsAt) {
    const sec = Math.max(0, Math.floor(state.endsAt - Date.now() / 1000));
    timerEl.textContent = `⏰ ${sec}s`;
  } else {
    timerEl.textContent = '';
  }
  roundEl.textContent = state.round ? `Manche ${state.round}/${state.totalRounds}` : '';
  letterEl.textContent = state.letter ? `Lettre : ${state.letter}` : '';
}

function handlePhaseDisplay() {
  if (renderedPhase !== state.phase) {
    renderedPhase = state.phase;
    
    answersForm.style.display = (state.phase === "playing") ? "" : "none";
    votesForm.style.display = (state.phase === "validation") ? "" : "none";
    scoreboard.style.display = (state.phase === "finished") ? "" : "none";

    if (state.phase === "playing") statusEl.textContent = "À vos claviers !";
    else if (state.phase === "validation") statusEl.textContent = "Phase de vote";
    else if (state.phase === "finished") statusEl.textContent = "Partie terminée !";
    else statusEl.textContent = "En attente...";

    // Sauvegarde forcée au changement de phase
    if (state.phase !== "playing" && Object.keys(localAnswers).length > 0) {
      sendAnswers(); 
    }
  }
}

// --- LOGIQUE INPUTS (FIX FOCUS + AUTO-SAVE) ---

function renderCategoriesOnce() {
  // 1. Protection absolue contre le re-render
  // Si on a déjà des inputs et que c'est la même manche, ON STOPPE TOUT.
  if (renderedRound === state.round && categoriesDiv.children.length > 0) {
    return; 
  }

  console.log("Construction des inputs pour la manche " + state.round);
  renderedRound = state.round;
  categoriesDiv.innerHTML = ""; 
  localAnswers = {}; 

  // Récupérer les réponses que le serveur connaît déjà (utile après un refresh page)
  const serverAnswers = (state.answers && state.answers[state.players.find(p => p.UserID === state.userID)?.UserID]) || {};
  // Note: Dans ta structure Go, StateForUser renvoie 'answers' qui contient TA réponse à toi.
  // Adapte ici selon la structure exacte du JSON reçu, mais généralement :
  const myAnswers = (state.answers && state.answers[state.userID]) ? state.answers[state.userID] : {};

  state.categories.forEach(cat => {
    // Priorité : ce qu'on a en mémoire > serveur > vide
    const val = localAnswers[cat.ID] || myAnswers[cat.ID] || "";
    localAnswers[cat.ID] = val; 

    const group = document.createElement('div');
    group.className = "form-group";
    
    const label = document.createElement('label');
    label.htmlFor = `cat_${cat.ID}`;
    label.textContent = cat.Name;

    const input = document.createElement('input');
    input.type = "text";
    input.id = `cat_${cat.ID}`;
    input.name = cat.ID;
    input.value = val;
    input.autocomplete = "off";

    // AUTO-SAVE : Quand on tape, on met à jour localAnswers et on lance un timer
    input.addEventListener('input', (e) => {
      localAnswers[cat.ID] = e.target.value;
      triggerAutoSave();
    });

    group.appendChild(label);
    group.appendChild(input);
    categoriesDiv.appendChild(group);
  });
}

function triggerAutoSave() {
  if (debounceTimer) clearTimeout(debounceTimer);
  debounceTimer = setTimeout(() => {
    console.log("Auto-save...");
    sendAnswers();
  }, 2000); // Sauvegarde auto toutes les 2 secondes d'inactivité
}

// --- LOGIQUE VOTES (FIX PERSISTANCE) ---

function renderVotesSmart() {
  if (!state.categories || !state.players || !state.answers) return;

  state.players.forEach(player => {
    if (player.UserID === state.userID) return; // On ne vote pas pour soi
    
    // Vérifier si le joueur a des réponses
    const playerAnswers = state.answers[player.UserID];
    if (!playerAnswers) return;

    const playerFieldsetID = `fieldset_p_${player.UserID}`;
    let fieldset = document.getElementById(playerFieldsetID);

    if (!fieldset) {
      fieldset = document.createElement('fieldset');
      fieldset.id = playerFieldsetID;
      fieldset.innerHTML = `<legend>${player.Pseudo}</legend>`;
      votesDiv.appendChild(fieldset);
    }

    state.categories.forEach(cat => {
      const answerText = playerAnswers[cat.ID];
      // On n'affiche pas les réponses vides
      if (!answerText) return;

      const uniqueID = `vote_${player.UserID}_${cat.ID}`;
      let voteRow = document.getElementById(`row_${uniqueID}`);

      // --- RECUPERATION DE LA VALEUR ACTUELLE DU VOTE ---
      // 1. Regarder si on a une action locale en attente
      let currentVal = localVotes[`${player.UserID}__${cat.ID}`];
      
      // 2. Sinon, regarder ce que le serveur a enregistré (si dispo)
      // Structure Go : votes[catID][targetUserID][voterID] => bool
      if (currentVal === undefined && state.votes && state.votes[cat.ID] && state.votes[cat.ID][player.UserID]) {
         const serverBool = state.votes[cat.ID][player.UserID][state.userID];
         if (serverBool === true) currentVal = "1";
         else if (serverBool === false) currentVal = "0";
      }

      if (currentVal === undefined) currentVal = ""; // Valeur par défaut

      if (!voteRow) {
        // Création
        voteRow = document.createElement('div');
        voteRow.className = "form-group";
        voteRow.id = `row_${uniqueID}`;
        
        voteRow.innerHTML = `
          <label for="${uniqueID}">
            ${cat.Name} : <strong style="color:#4f8cff">${answerText}</strong>
          </label>
          <select id="${uniqueID}" data-player="${player.UserID}" data-cat="${cat.ID}">
            <option value="">Juger...</option>
            <option value="1">✅ Valide</option>
            <option value="0">❌ Refusé</option>
          </select>
        `;
        fieldset.appendChild(voteRow);

        const select = voteRow.querySelector('select');
        select.value = currentVal; // Appliquer la valeur
        
        select.addEventListener('change', (e) => {
           const key = `${player.UserID}__${cat.ID}`;
           localVotes[key] = e.target.value;
           // Envoi immédiat du vote pour fluidité (optionnel, mais mieux)
           // sendVotes(); 
        });
      } else {
        // MISE A JOUR : Si l'élément existe, on vérifie si on doit mettre à jour la sélection
        // On le fait seulement si l'utilisateur n'est pas en train d'interagir avec (compliqué à détecter)
        // ou simplement, on force la valeur si elle vient du serveur pour confirmer
        const select = voteRow.querySelector('select');
        if (select.value === "" && currentVal !== "") {
            select.value = currentVal;
        }
      }
    });
  });
}



// --- COMMUNICATION ---

function sendAnswers() {
  const data = {};
  state.categories.forEach(cat => {
      data[cat.ID] = localAnswers[cat.ID] || "";
  });
  
  fetch(`/api/salle/${roomCode}/petitbac/answers`, {
    method: "POST",
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify(data)
  }).then(() => console.log("Réponses sync server OK"));
}

answersForm.onsubmit = function(e) {
  e.preventDefault();
  sendAnswers();
  statusEl.textContent = "Réponses envoyées !";
};

function sendVotes() {
  const data = {};
  Object.entries(localVotes).forEach(([key, val]) => {
    if (val === "") return;
    const [pID, cID] = key.split("__");
    if (!data[cID]) data[cID] = {};
    data[cID][pID] = (val === "1");
  });

  fetch(`/api/salle/${roomCode}/petitbac/votes`, {
    method: "POST",
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify(data)
  }).then(() => {
      statusEl.textContent = "Votes pris en compte !";
      // On ne vide pas localVotes ici pour garder l'état visuel
  });
}

votesForm.onsubmit = function(e) {
  e.preventDefault();
  sendVotes();
};

// --- WEBSOCKET ---

function connectWS() {
  const proto = location.protocol === "https:" ? "wss" : "ws";
  ws = new WebSocket(`${proto}://${location.host}/ws/salle/${encodeURIComponent(roomCode)}`);
  
  ws.onopen = () => {
    console.log("WS Connecté");
    fetchState();
  };
  
  ws.onmessage = () => {
      fetchState(); 
  };
  
  ws.onclose = () => {
      setTimeout(connectWS, 2000);
  };
}

connectWS();
setInterval(fetchState, 5000); // Polling de sécurité
setInterval(updateHeaderInfo, 1000); // Timer visuel