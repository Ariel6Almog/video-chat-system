import React, { useEffect, useRef, useState } from "react";
import "./Chat.css";

export default function Chat() {
  const videoRef = useRef(null);
  const [err, setErr] = useState("");

  useEffect(() => {
    let localStream;

    (async () => {
      try {
        localStream = await navigator.mediaDevices.getUserMedia({
          video: { width: { ideal: 1280 }, height: { ideal: 720 } },
          audio: true,
        });

        if (videoRef.current) {
          videoRef.current.srcObject = localStream;
          videoRef.current.play().catch(() => {});
        }
      } catch (e) {
        setErr(e?.message || "Failed to access camera/mic");
      }
    })();

    return () => {
      localStream?.getTracks().forEach(t => t.stop());
    };
  }, []);

  return (
    <div className="chat-container">
      <h2>Video Chat Room</h2>

      <div className="video-grid">
        <div className="video-box">
          <p>You</p>
          <video
            className="video"
            ref={videoRef}
            playsInline
            autoPlay
            muted
          ></video>
        </div>

        <div className="video-box">
          <p>Other User</p>
          <video className="video" playsInline autoPlay></video>
        </div>
      </div>

      <div className="controls">
        <button className="end-call">End Call</button>
        <button className="mute">Mute</button>
        <button className="camera">Toggle Camera</button>
      </div>

      {err && <div className="error">{err}</div>}
    </div>
  );
}
