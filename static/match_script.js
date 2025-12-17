(function () {
  const code = document.body?.dataset?.roomCode;
  if (!code) return;

  const audio = document.getElementById("audio");
  const statusEl = document.getElementById("status");
  const timerEl = document.getElementById("timer");
  const revealEl = document.getElementById("reveal");
  const scoreboard = document.getElementById("scoreboard");
  // scoreList est désormais géré par scoreboard_render.js
  const replayBtn = document.getElementById("replayBtn");
  const form = document.getElementById("guessForm");
  const guessInput = document.getElementById("guess");

  let endsAtUnix = 0;
  let phase = "idle";
  
  // --- NOUVEAU : Initialisation de l'état global pour le scoreboard externe ---
  window.state = {
    players: [],
    userID: 0, // Sera rempli si l'API le renvoie
    phase: "idle"
  };

  function api(path) {
    return `/api/salle/${encodeURIComponent(code)}/blindtest/${path}`;
  }
  function playersApi() {
    return `/api/salle/${encodeURIComponent(code)}/players`;
  }

  function setPhase(p) {
    phase = p || "idle";
    // Synchro avec l'état global pour le renderer
    window.state.phase = phase;
  }

  function startTimerUI() {
    function tick() {
      if (!endsAtUnix) {
        timerEl.textContent = "";
        return;
      }
      const ms = (endsAtUnix * 1000) - Date.now();
      const s = Math.max(0, Math.floor(ms / 1000));
      timerEl.textContent = `Temps restant : ${s}s`;
      if (ms > 0) requestAnimationFrame(tick);
    }
    requestAnimationFrame(tick);
  }

  function playPreviewLoop(previewUrl) {
    if (!previewUrl) return;
    audio.src = previewUrl;
    audio.currentTime = 0;
    audio.play().catch(() => {});

    audio.onended = () => {
      if (phase !== "playing") return;
      if (!endsAtUnix) return;
      if (Date.now() < endsAtUnix * 1000) {
        audio.currentTime = 0;
        audio.play().catch(() => {});
      }
    };
  }

  async function refreshState() {
    const res = await fetch(api("state"));
    const st = await res.json();

    setPhase(st.phase);
    // Si l'API renvoie l'ID utilisateur, on le stocke pour le surlignage "Moi"
    if (st.userID) window.state.userID = st.userID;

    currentRound = st.round || 0;
    totalRounds = st.total_rounds || 0;
    endsAtUnix = st.ends_at_unix || 0;

    if (phase === "idle") {
      statusEl.textContent = "En attente du lancement…";
      form.style.display = "none";
      revealEl.textContent = "";
      audio.pause();
      if (scoreboard) scoreboard.style.display = "none";
      return;
    }

    if (phase === "playing") {
      statusEl.textContent = `Manche ${st.round}/${st.total_rounds}`;
      revealEl.textContent = "";
      form.style.display = "";
      guessInput.disabled = !!st.already_tried;
      if (st.preview_url) playPreviewLoop(st.preview_url);
      startTimerUI();
      if (scoreboard) scoreboard.style.display = "none";
      return;
    }

    if (phase === "reveal" || phase === "finished") {
      statusEl.textContent = phase === "finished" ? "Partie terminée" : `Révélation (${st.round}/${st.total_rounds})`;
      form.style.display = phase === "finished" ? "none" : "";
      guessInput.disabled = true;
      audio.pause();
      if (st.title || st.artist) {
        revealEl.textContent = `Réponse : ${st.title || ""} — ${st.artist || ""}`;
      }
      
      // CHARGEMENT DU SCOREBOARD VIA LE NOUVEAU SYSTÈME
      await loadAndRenderScoreboard(true);
      return;
    }
  }

  // --- NOUVELLE FONCTION DE CHARGEMENT ---
  async function loadAndRenderScoreboard(show) {
    if (!scoreboard) return;
    try {
      const res = await fetch(playersApi());
      if (!res.ok) return;
      
      const players = await res.json();
      
      // 1. Mettre à jour les données globales
      window.state.players = players;
      
      // 2. Appeler le moteur de rendu externe (scoreboard_render.js)
      if (typeof renderScoreboard === "function") {
        renderScoreboard(); 
      }

      // 3. Afficher la section
      if (show) scoreboard.style.display = "";
    } catch (e) {
      console.error("Erreur chargement scoreboard:", e);
    }
  }

  const proto = (location.protocol === "https:") ? "wss" : "ws";
  const ws = new WebSocket(`${proto}://${location.host}/ws/salle/${encodeURIComponent(code)}`);

  ws.onopen = () => refreshState().catch(() => {});
  ws.onmessage = (ev) => {
    try {
      const msg = JSON.parse(ev.data);

      if (msg.type === "blindtest_round_started") {
        setPhase("playing");
        statusEl.textContent = `Manche ${msg.payload.round}/${msg.payload.total_rounds}`;
        revealEl.textContent = "";
        endsAtUnix = msg.payload.ends_at_unix;
        guessInput.disabled = false;
        guessInput.value = "";
        if (scoreboard) scoreboard.style.display = "none";
        playPreviewLoop(msg.payload.preview_url);
        startTimerUI();
        return;
      }

      if (msg.type === "blindtest_round_reveal") {
        setPhase("reveal");
        audio.pause();
        guessInput.disabled = true;
        statusEl.textContent = "Révélation…";
        revealEl.textContent = `Réponse : ${msg.payload.title} — ${msg.payload.artist}`;
        // Afficher les scores à la révélation
        loadAndRenderScoreboard(true);
        return;
      }

      if (msg.type === "blindtest_finished") {
        setPhase("finished");
        endsAtUnix = 0;
        timerEl.textContent = "";
        audio.pause();
        form.style.display = "none";
        statusEl.textContent = "Partie terminée";
        loadAndRenderScoreboard(true).catch(() => {});
        return;
      }

      if (msg.type === "room_updated" && (phase === "finished" || phase === "reveal")) {
        loadAndRenderScoreboard(true).catch(() => {});
        return;
      }
    } catch (_) {}
  };

  form.addEventListener("submit", async (e) => {
    e.preventDefault();
    const guess = guessInput.value || "";

    const res = await fetch(api("guess"), {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ guess })
    });
    const out = await res.json();

    if (out.locked || out.already_tried) {
      guessInput.disabled = true;
    }
    if (out.correct) {
      statusEl.textContent = `Bonne réponse ! +${out.points_awarded} pts`;
      guessInput.disabled = true;
    } else if (!out.locked) {
      statusEl.textContent = "Raté pour cette manche.";
    }
  });

  if (replayBtn) {
    replayBtn.addEventListener("click", () => {
      window.location.href = `/salle/${encodeURIComponent(code)}`;
    });
  }

  // fallback si WS n'est pas connecté
  setInterval(() => {
    if (ws.readyState !== 1) refreshState().catch(() => {});
  }, 1500);
})();