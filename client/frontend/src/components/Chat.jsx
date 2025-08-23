import React, { useEffect, useMemo, useRef, useState } from "react";


function genUUID() {
  return ([1e7]+-1e3+-4e3+-8e3+-1e11).replace(/[018]/g, c =>
    (c ^ crypto.getRandomValues(new Uint8Array(1))[0] & 15 >> c / 4).toString(16)
  );
}

export default function Chat({ sessionId, token, dasherBase = "http://localhost:8090", showRecordedPreview = false }) {
  const format = pickMime();
  const previewRef = useRef(null);
  const playbackRef = useRef(null);
  const [stream, setStream] = useState(null);
  const [rec, setRec] = useState(null);
  const [chunks, setChunks] = useState([]);
  const [status, setStatus] = useState("idle");
  const [err, setErr] = useState("");
  const [wsConn, setWsConn] = useState(null);
  const [isPublishing, setIsPublishing] = useState(false);
  const [cams, setCams] = useState([]);
  const [mics, setMics] = useState([]);
  const [camId, setCamId] = useState("");
  const [micId, setMicId] = useState("");
  const [publisherId] = useState(() => genUUID());


  const wsUrl = useMemo(() => buildIngestUrl(sessionId, token, publisherId), [sessionId, token, publisherId]);

  useEffect(() => {
    let mounted = true;
    (async () => {
      try {
        const { cameras, microphones } = await listDevices();
        if (!mounted) return;
        setCams(cameras);
        setMics(microphones);
        setCamId(cameras[0]?.deviceId || "");
        setMicId(microphones[0]?.deviceId || "");
      } catch (e) {
        setErr(e?.message || "enumerateDevices failed");
      }
    })();
    return () => { mounted = false; };
  }, []);

  useEffect(() => {
    let s;
    (async () => {
      try {
        s = await openUserMedia(camId, micId);
        setStream(s);
        if (previewRef.current) {
          previewRef.current.srcObject = s;
          await previewRef.current.play().catch(() => {});
        }
      } catch (e) {
        setErr(e?.message || "getUserMedia failed");
      }
    })();
    return () => {
      try { previewRef.current && (previewRef.current.srcObject = null); } catch {}
      s?.getTracks().forEach(t => t.stop());
    };
  }, [camId, micId]);

  const startRecord = () => {
    if (!stream) return;
    const tmp = [];
    const r = new MediaRecorder(stream, { mimeType: format });
    r.ondataavailable = ev => { if (ev.data?.size) tmp.push(ev.data); };
    r.onstop = () => {
      const blob = new Blob(tmp, { type: format });
      const url = URL.createObjectURL(blob);
      if (playbackRef.current) {
        playbackRef.current.src = url;
        playbackRef.current.play().catch(() => {});
      }
      setChunks(tmp);
      setStatus("stopped");
    };
    r.start(1000);
    setRec(r);
    setChunks([]);
    setStatus("recording");
  };

  const stopRecord = () => {
    if (rec?.state !== "inactive") rec.stop();
  };

  const startPublish = () => {
    if (!stream || isPublishing) return;
    connectWsWithRetry(wsUrl, stream, format, setRec, setIsPublishing, setStatus, setErr, setWsConn);
  };

  const stopPublish = () => {
    if (rec?.state !== "inactive") rec.stop();
    try { wsConn?.close(); } catch {}
    setIsPublishing(false);
    setStatus("idle");
  };

  return (
    <div className="chat-container">
      <h2>Video Chat Room</h2>

      <div style={{ display: "flex", gap: 8, marginBottom: 12 }}>
        <select value={camId} onChange={e => setCamId(e.target.value)}>
          {cams.map(c => <option key={c.deviceId} value={c.deviceId}>{c.label || "Camera"}</option>)}
        </select>
        <select value={micId} onChange={e => setMicId(e.target.value)}>
          {mics.map(m => <option key={m.deviceId} value={m.deviceId}>{m.label || "Microphone"}</option>)}
        </select>
      </div>

      <div className="video-grid">
        <div className="video-box">
          <p>You</p>
          <video className="video" ref={previewRef} playsInline autoPlay muted />
        </div>

        {showRecordedPreview && (
          <div className="video-box">
            <p>Recorded Preview</p>
            <video className="video" ref={playbackRef} playsInline controls />
          </div>
        )}
      </div>

      <div className="controls">
        {showRecordedPreview && (
          <>
            <button onClick={startRecord} disabled={!stream || status === "recording"}>Start Rec</button>
            <button onClick={stopRecord} disabled={status !== "recording"}>Stop Rec</button>
          </>
        )}
        <button onClick={startPublish} disabled={!stream || isPublishing}>Start Publish</button>
        <button onClick={stopPublish} disabled={!isPublishing}>Stop Publish</button>
      </div>

      <Viewer sessionId={sessionId} dasherBase={dasherBase} />

      {err && <div className="error">{err}</div>}
    </div>
  );
}

function pickMime() {
  const cands = ["video/webm;codecs=h264,opus", "video/webm;codecs=vp8,opus", "video/webm"];
  return cands.find(t => window.MediaRecorder && MediaRecorder.isTypeSupported(t)) || "video/webm";
}

async function listDevices() {
  try { await navigator.mediaDevices.getUserMedia({ audio: true, video: true }); } catch {}
  const all = await navigator.mediaDevices.enumerateDevices();
  const cameras = all.filter(d => d.kind === "videoinput");
  const microphones = all.filter(d => d.kind === "audioinput");
  return { cameras, microphones };
}

async function openUserMedia(camId, micId) {
  const constraints = {
    video: camId ? { deviceId: { exact: camId }, width: 1280, height: 720, frameRate: { ideal: 30 } } : true,
    audio: micId ? { deviceId: { exact: micId } } : true
  };
  return await navigator.mediaDevices.getUserMedia(constraints);
}

function buildIngestUrl(sessionId, token, publisherId) {
  const override = (import.meta.env.VITE_INGEST_URL || "").trim();
  if (override) {
    const scheme = override.startsWith("ws") ? "" : (window.location.protocol === "https:" ? "wss://" : "ws://");
    const base = scheme ? scheme + override.replace(/^\/\//, "") : override;
    return `${base.replace(/\/$/, "")}/ws/ingest/${sessionId}/${publisherId}?token=${encodeURIComponent(token)}`;
  }
  const isSecure = window.location.protocol === "https:";
  const scheme = isSecure ? "wss" : "ws";
  const host = window.location.hostname || "localhost";
  const port = window.location.port || (isSecure ? "443" : "8080");
  return `${scheme}://${host}:${port}/ws/ingest/${sessionId}/${publisherId}?token=${encodeURIComponent(token)}`;
}



function connectWsWithRetry(wsUrl, stream, mimeType, setRec, setIsPublishing, setStatus, setErr, setWsConn) {
  let tries = 0;
  const maxTries = 5;
  const backoff = () => Math.min(1000 * Math.pow(2, tries), 10000);

  const connect = () => {
    const ws = new WebSocket(wsUrl);
    ws.binaryType = "arraybuffer";
    setWsConn(ws);

    let keepalive;
    let r;

    ws.onopen = () => {
      r = new MediaRecorder(stream, { mimeType });
      r.ondataavailable = async ev => {
        if (!ev.data || !ev.data.size || ws.readyState !== 1) return;
        const buf = await ev.data.arrayBuffer();
        if (ws.bufferedAmount > 8 * 1024 * 1024) {
          r.pause();
          const i = setInterval(() => {
            if (ws.bufferedAmount < 1 * 1024 * 1024) {
              clearInterval(i);
              r.resume();
            }
          }, 100);
        }
        ws.send(buf);
      };
      r.start(1000);
      setRec(r);
      setIsPublishing(true);
      setStatus("publishing");
      keepalive = setInterval(() => {
        try { ws.readyState === 1 && ws.send("ping"); } catch {}
      }, 15000);
      tries = 0;
    };

    ws.onclose = () => {
      try { clearInterval(keepalive); } catch {}
      try { r && r.state !== "inactive" && r.stop(); } catch {}
      setIsPublishing(false);
      setStatus("closed");
      if (tries < maxTries) {
        tries += 1;
        setTimeout(connect, backoff());
      } else {
        setErr("WebSocket closed");
      }
    };

    ws.onerror = () => {
      setErr("WebSocket error");
    };
  };

  connect();
}

function Viewer({ sessionId, dasherBase }) {
  const playerRef = useRef(null);
  const dashRef = useRef(null);
  const [ready, setReady] = useState(false);

  useEffect(() => {
    const script = document.createElement("script");
    script.src = "https://cdn.dashjs.org/latest/dash.all.min.js";
    script.onload = () => setReady(true);
    document.body.appendChild(script);
    return () => { document.body.removeChild(script); };
  }, []);

  useEffect(() => {
    if (!ready || !playerRef.current || !window.dashjs) return;
    const url = `${dasherBase}/dash/${sessionId}/manifest.mpd`;
    dashRef.current = window.dashjs.MediaPlayer().create();
    dashRef.current.initialize(playerRef.current, url, true);
    return () => {
      try { dashRef.current?.reset(); } catch {}
    };
  }, [ready, sessionId, dasherBase]);

  return (
    <div style={{ marginTop: 16 }}>
      <p>Viewer (DASH)</p>
      <video ref={playerRef} controls playsInline style={{ width: 400, background: "#000" }} />
    </div>
  );
}
