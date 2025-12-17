const roomCode = document.body.dataset.roomCode;
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
let debounceTimer = null; 

let localAnswers = {}; 
let localVotes = {};   

let renderedRound = -1;
let renderedPhase = null;
let renderedVoteRound = -1; 


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

    if (state.phase !== "playing" && Object.keys(localAnswers).length > 0) {
      sendAnswers(); 
    }
  }
}


function renderCategoriesOnce() {

  if (renderedRound === state.round && categoriesDiv.children.length > 0) {
    return; 
  }

  console.log("Construction des inputs pour la manche " + state.round);
  renderedRound = state.round;
  categoriesDiv.innerHTML = ""; 
  localAnswers = {}; 

  const serverAnswers = (state.answers && state.answers[state.players.find(p => p.UserID === state.userID)?.UserID]) || {};

  const myAnswers = (state.answers && state.answers[state.userID]) ? state.answers[state.userID] : {};

  state.categories.forEach(cat => {

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
  }, 2000); 
}


function renderVotesSmart() {
  if (!state.categories || !state.players || !state.answers) return;
  if (renderedVoteRound !== state.round) {
    console.log("Nouvelle manche de vote détectée (" + state.round + "). Nettoyage.");
    votesDiv.innerHTML = ""; 
    localVotes = {};        
    renderedVoteRound = state.round;
  }

  state.players.forEach(player => {
    if (player.UserID === state.userID) return; 
    
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
      if (!answerText) return;

      const uniqueID = `vote_${player.UserID}_${cat.ID}`;
      let voteRow = document.getElementById(`row_${uniqueID}`);

      let currentVal = localVotes[`${player.UserID}__${cat.ID}`];
      
     
      if (currentVal === undefined && state.votes && state.votes[cat.ID] && state.votes[cat.ID][player.UserID]) {
         const serverBool = state.votes[cat.ID][player.UserID][state.userID];
         if (serverBool === true) currentVal = "1";
         else if (serverBool === false) currentVal = "0";
      }

      if (currentVal === undefined) currentVal = ""; 

      if (!voteRow) {
       
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
        select.value = currentVal; 
        
        select.addEventListener('change', (e) => {
           const key = `${player.UserID}__${cat.ID}`;
           localVotes[key] = e.target.value;
        });
      } else {
     
        const select = voteRow.querySelector('select');
        if (select.value === "" && currentVal !== "") {
            select.value = currentVal;
        }
      }
    });
  });
}


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
      
  });
}

votesForm.onsubmit = function(e) {
  e.preventDefault();
  sendVotes();
};



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
setInterval(fetchState, 5000); 
setInterval(updateHeaderInfo, 1000); 