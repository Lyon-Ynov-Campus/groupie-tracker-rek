(function () {
  const code = document.body?.dataset?.roomCode;
  if (!code) return;

  const statusEl = document.getElementById("status");
  const timerEl = document.getElementById("timer");
  const letterEl = document.getElementById("letter");
  const answersForm = document.getElementById("answersForm");
  const categoriesDiv = document.getElementById("categories");
  const votesForm = document.getElementById("votesForm");
  const votesDiv = document.getElementById("votes");
  const scoreboard = document.getElementById("scoreboard");
  const scoreList = document.getElementById("scoreList");
  const replayBtn = document.getElementById("replayBtn");

  let phase = "idle";
  let endsAtUnix = 0;
  let categories = [];
  let letter = "";
  let answers = {};
  let allAnswers = {};
  let votes = {};

  function api(path, opts) {
    return fetch(`/api/salle/${encodeURIComponent(code)}/petitbac/${path}`, opts);
  }

  function show(el) { el && (el.style.display = ""); }
  function hide(el) { el && (el.style.display = "none"); }

  function renderCategories(cats, prevAnswers = {}) {
    categoriesDiv.innerHTML = "";
    cats.forEach(cat => {
      const group = document.createElement("div");
      group.className = "form-group";
      const label = document.createElement("label");
      label.htmlFor = "cat" + cat.ID;
      label.textContent = cat.Name;
      const input = document.createElement("input");
      input.type = "text";
      input.id = "cat" + cat.ID;
      input.name = "cat" + cat.ID;
      input.autocomplete = "off";
      input.value = prevAnswers[cat.ID] || "";
      group.appendChild(label);
      group.appendChild(input);
      categoriesDiv.appendChild(group);
    });
  }

  function renderVotes(votesData, cats, players) {
    votesDiv.innerHTML = "";
    cats.forEach(cat => {
      const catDiv = document.createElement("div");
      catDiv.className = "form-group";
      const catLabel = document.createElement("label");
      catLabel.textContent = cat.Name;
      catDiv.appendChild(catLabel);

      const ul = document.createElement("ul");
      ul.style.listStyle = "none";
      ul.style.padding = 0;

      (votesData[cat.ID] || []).forEach(ans => {
        const li = document.createElement("li");
        const chk = document.createElement("input");
        chk.type = "checkbox";
        chk.name = `vote_${cat.ID}_${ans.userID}`;
        chk.checked = ans.valid;
        li.appendChild(chk);
        li.appendChild(document.createTextNode(` ${ans.pseudo}: ${ans.answer}`));
        ul.appendChild(li);
      });

      catDiv.appendChild(ul);
      votesDiv.appendChild(catDiv);
    });
  }

  function renderScoreboard(scores) {
    scoreList.innerHTML = "";
    scores.forEach(s => {
      const li = document.createElement("li");
      li.textContent = `${s.pseudo} : ${s.score}`;
      scoreList.appendChild(li);
    });
  }

  function startTimerUI() {
    if (!endsAtUnix) { timerEl.textContent = ""; return; }
    function update() {
      const now = Math.floor(Date.now() / 1000);
      const left = Math.max(0, endsAtUnix - now);
      timerEl.textContent = left > 0 ? `Temps restant : ${left}s` : "";
      if (left > 0) setTimeout(update, 500);
    }
    update();
  }

  async function refreshState() {
    const res = await api("state");
    if (!res.ok) { statusEl.textContent = "Erreur de connexion."; return; }
    const state = await res.json();
    phase = state.phase;
    endsAtUnix = state.endsAt;
    letter = state.letter || "";
    categories = state.categories || [];
    answers = state.answers || {};
    allAnswers = state.allAnswers || {};
    votes = state.votes || {};
    const scores = state.scores || [];
    const players = state.players || [];

    letterEl.textContent = letter ? `Lettre : ${letter}` : "";

    // Afficher le round si présent
    if (state.round !== undefined && state.totalRounds !== undefined) {
      document.getElementById("round").textContent =
        "Manche " + state.round + " / " + state.totalRounds;
    }

    if (phase === "playing") {
      statusEl.textContent = "À toi de jouer !";
      show(answersForm);
      hide(votesForm);
      hide(scoreboard);
      renderCategories(categories, answers);
      startTimerUI();
    } else if (phase === "validation") {
      statusEl.textContent = "Vote sur les réponses des autres joueurs.";
      hide(answersForm);
      show(votesForm);
      hide(scoreboard);
      renderVotes(allAnswers, categories, players);
      startTimerUI();
    } else if (phase === "finished") {
      statusEl.textContent = "Partie terminée !";
      hide(answersForm);
      hide(votesForm);
      show(scoreboard);
      renderScoreboard(scores);
      timerEl.textContent = "";
    } else {
      statusEl.textContent = "En attente du début de la partie…";
      hide(answersForm);
      hide(votesForm);
      hide(scoreboard);
      timerEl.textContent = "";
      letterEl.textContent = "";
    }
  }

  answersForm.addEventListener("submit", async (e) => {
    e.preventDefault();
    const data = {};
    categories.forEach(cat => {
      const val = document.getElementById("cat" + cat.ID)?.value || "";
      data[cat.ID] = val;
    });
    await api("answers", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(data)
    });
    await refreshState();
  });

  votesForm.addEventListener("submit", async (e) => {
    e.preventDefault();
    const data = {};
    categories.forEach(cat => {
      data[cat.ID] = {};
      const catVotes = votesDiv.querySelectorAll(`input[name^="vote_${cat.ID}_"]`);
      catVotes.forEach(input => {
        const userID = input.name.split("_")[2];
        data[cat.ID][userID] = input.checked;
      });
    });
    await api("votes", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(data)
    });
    await refreshState();
  });

  if (replayBtn) {
    replayBtn.addEventListener("click", async () => {
      await api("restart", { method: "POST" });
      await refreshState();
    });
  }

  // WebSocket pour rafraîchir l’état
  const proto = (location.protocol === "https:") ? "wss" : "ws";
  const ws = new WebSocket(`${proto}://${location.host}/ws/salle/${encodeURIComponent(code)}`);
  ws.onmessage = () => refreshState();

  refreshState();
})();