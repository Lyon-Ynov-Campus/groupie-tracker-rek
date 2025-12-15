(function () {
  const code = document.body?.dataset?.roomCode;
  if (!code) return;

  const proto = (location.protocol === "https:") ? "wss" : "ws";
  const ws = new WebSocket(`${proto}://${location.host}/ws/salle/${encodeURIComponent(code)}`);

  ws.onmessage = (ev) => {
    try {
      const msg = JSON.parse(ev.data);
      if (msg.type === "room_updated") {
        // léger rafraîchissement d'UI (animation possible avant reload)
        location.reload();
        return;
      }
      // redirection client vers l'écran de jeu quand le serveur indique qu'une manche démarre
      if (msg.type === "blindtest_round_started" || msg.type === "petitbac_round_started") {
        location.href = `/game/${encodeURIComponent(code)}`;
        return;
      }
    } catch (e) {
      // ignore malformed messages
    }
  };
})();