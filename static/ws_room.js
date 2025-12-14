(function () {
  const code = document.body?.dataset?.roomCode;
  if (!code) return;

  const proto = (location.protocol === "https:") ? "wss" : "ws";
  const ws = new WebSocket(`${proto}://${location.host}/ws/salle/${encodeURIComponent(code)}`);

  ws.onmessage = (ev) => {
    try {
      const msg = JSON.parse(ev.data);
      if (msg.type === "room_updated") {
        location.reload();
      }
    } catch (_) {}
  };
})();