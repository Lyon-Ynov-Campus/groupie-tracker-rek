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
let lastPhase = null;
let localAnswers = {};

function fetchState() {
  fetch(`/api/salle/${roomCode}/petitbac/state`)
    .then(r => r.json())
    .then(updateUI)
    .catch(() => statusEl.textContent = "Erreur de connexion.");
}

function updateUI(newState) {
  state = newState;
  if (!state || !state.phase) {
    statusEl.textContent = "En attente du jeu…";
    return;
  }
  // Timer & round
  timerEl.textContent = state.endsAt ? `⏰ ${Math.max(0, Math.floor(state.endsAt - Date.now() / 1000))}s` : '';
  roundEl.textContent = state.round ? `Manche ${state.round}` : '';
  letterEl.textContent = state.letter ? `Lettre : ${state.letter}` : '';

  if (state.phase === "playing") {
    statusEl.textContent = "Remplis tes réponses !";
    answersForm.style.display = "";
    votesForm.style.display = "none";
    scoreboard.style.display = "none";
    renderCategories();
  } else if (state.phase === "validation") {
    statusEl.textContent = "Vote sur les réponses des autres !";
    answersForm.style.display = "none";
    votesForm.style.display = "";
    scoreboard.style.display = "none";
    renderVotes();
  } else if (state.phase === "finished") {
    statusEl.textContent = "Scores finaux";
    answersForm.style.display = "none";
    votesForm.style.display = "none";
    scoreboard.style.display = "";
    renderScoreboard();
  } else {
    statusEl.textContent = "En attente des joueurs…";
    answersForm.style.display = "none";
    votesForm.style.display = "none";
    scoreboard.style.display = "none";
  }

  // Soumission automatique si on quitte la phase "playing" sans avoir validé
  if (lastPhase === "playing" && newState.phase !== "playing") {
    if (answersForm.style.display !== "none") {
      const data = {};
      state.categories.forEach(cat => {
        data[cat.ID] = localAnswers[cat.ID] || "";
      });
      fetch(`/api/salle/${roomCode}/petitbac/answers`, {
        method: "POST",
        headers: {'Content-Type': 'application/json'},
        body: JSON.stringify(data)
      });
      localAnswers = {};
    }
  }
  lastPhase = newState.phase;
}

function renderCategories() {
  categoriesDiv.innerHTML = "";
  if (!state.categories) return;

  // Reset localAnswers si on change de manche
  if (!localAnswers.__round || localAnswers.__round !== state.round) {
    localAnswers = {__round: state.round};
    const myAnswers = (state.answers && state.answers[state.userID]) || {};
    state.categories.forEach(cat => {
      localAnswers[cat.ID] = myAnswers[cat.ID] || "";
    });
  }

  state.categories.forEach(cat => {
    const val = localAnswers[cat.ID] || "";
    const div = document.createElement('div');
    div.className = "form-group";
    div.innerHTML = `
      <label for="cat${cat.ID}">${cat.Name}</label>
      <input type="text" id="cat${cat.ID}" name="${cat.ID}" value="${val}" autocomplete="off">
    `;
    // MAJ localAnswers à chaque frappe
    div.querySelector('input').addEventListener('input', e => {
      localAnswers[cat.ID] = e.target.value;
    });
    categoriesDiv.appendChild(div);
  });
}

function renderVotes() {
  votesDiv.innerHTML = "";
  if (!state.categories || !state.players || !state.answers) return;
  state.players.forEach(player => {
    if (!state.answers[player.UserID]) return;
    const group = document.createElement('fieldset');
    group.innerHTML = `<legend>${player.Pseudo}</legend>`;
    state.categories.forEach(cat => {
      const answer = state.answers[player.UserID][cat.ID] || "";
      const id = `vote_${player.UserID}_${cat.ID}`;
      group.innerHTML += `
        <div class="form-group">
          <label for="${id}">${cat.Name} : <b>${answer}</b></label>
          <select name="${player.UserID}__${cat.ID}" id="${id}" required>
            <option value="">Vote…</option>
            <option value="1">Valide</option>
            <option value="0">Refusé</option>
          </select>
        </div>
      `;
    });
    votesDiv.appendChild(group);
  });
}

function renderScoreboard() {
  scoreList.innerHTML = "";
  if (!state.players) return;
  state.players.forEach(player => {
    const li = document.createElement('li');
    li.textContent = `${player.Pseudo} : ${player.Score} pts`;
    scoreList.appendChild(li);
  });
}

// Form submissions
answersForm.onsubmit = function(e) {
  e.preventDefault();
  const data = {};
  state.categories.forEach(cat => {
    data[cat.ID] = localAnswers[cat.ID] || "";
  });
  fetch(`/api/salle/${roomCode}/petitbac/answers`, {
    method: "POST",
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify(data)
  }).then(() => {
    localAnswers = {};
    fetchState();
  });
};

votesForm.onsubmit = function(e) {
  e.preventDefault();
  const data = {};
  new FormData(votesForm).forEach((v, k) => {
    const [userID, catID] = k.split("__");
    if (!data[catID]) data[catID] = {};
    data[catID][userID] = v === "1";
  });
  fetch(`/api/salle/${roomCode}/petitbac/votes`, {
    method: "POST",
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify(data)
  }).then(fetchState);
};

// WebSocket for live updates
function connectWS() {
  const proto = location.protocol === "https:" ? "wss" : "ws";
  ws = new WebSocket(`${proto}://${location.host}/ws/salle/${encodeURIComponent(roomCode)}`);
  ws.onmessage = fetchState;
  ws.onclose = () => setTimeout(connectWS, 2000);
}
connectWS();
fetchState();
setInterval(fetchState, 5000);

function updateTimerUI() {
  if (state && state.endsAt) {
    const sec = Math.max(0, Math.floor(state.endsAt - Date.now() / 1000));
    timerEl.textContent = `⏰ ${sec}s`;
  }
}
setInterval(updateTimerUI, 1000);