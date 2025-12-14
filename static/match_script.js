(function () {
  const code = document.body?.dataset?.roomCode;
  if (!code) return;

  const audio = document.getElementById("audio");
  const statusEl = document.getElementById("status");
  const timerEl = document.getElementById("timer");
  const revealEl = document.getElementById("reveal");
  const form = document.getElementById("guessForm");
  const guessInput = document.getElementById("guess");

  let endsAtUnix = 0;
  let phase = "idle";

  function api(path) {
    return `/api/salle/${encodeURIComponent(code)}/blindtest/${path}`;
  }

  function setPhase(p) {
    phase = p || "idle";
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
    endsAtUnix = st.ends_at_unix || 0;

    if (phase === "idle") {
      statusEl.textContent = "En attente du lancement…";
      form.style.display = "none";
      revealEl.textContent = "";
      audio.pause();
      return;
    }

    if (phase === "playing") {
      statusEl.textContent = `Manche ${st.round}/${st.total_rounds}`;
      revealEl.textContent = "";
      form.style.display = "";
      guessInput.disabled = !!st.already_tried;
      if (st.preview_url) playPreviewLoop(st.preview_url);
      startTimerUI();
      return;
    }

    if (phase === "reveal" || phase === "finished") {
      form.style.display = "";
      guessInput.disabled = true;
      audio.pause();
      if (st.title || st.artist) {
        revealEl.textContent = `Réponse : ${st.title || ""} — ${st.artist || ""}`;
      }
      return;
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
        revealEl.textContent = "";
        endsAtUnix = msg.payload.ends_at_unix;
        guessInput.disabled = false;
        playPreviewLoop(msg.payload.preview_url);
        startTimerUI();
        return;
      }

      if (msg.type === "blindtest_round_reveal") {
        setPhase("reveal");
        audio.pause();
        guessInput.disabled = true;
        revealEl.textContent = `Réponse : ${msg.payload.title} — ${msg.payload.artist}`;
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
  });

  // fallback si WS n'est pas connecté
  setInterval(() => {
    if (ws.readyState !== 1) refreshState().catch(() => {});
  }, 1500);
})();