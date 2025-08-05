import React from "react";
import "./Chat.css";

function Chat() {
  return (
    <div className="chat-container">
      <h2>Video Chat Room</h2>

      <div className="video-grid">
        <div className="video-box">
          <p>You</p>
          <video className="video" autoPlay muted></video>
        </div>

        <div className="video-box">
          <p>Other User</p>
          <video className="video" autoPlay></video>
        </div>
      </div>

      <div className="controls">
        <button className="end-call">End Call</button>
        <button className="mute">Mute</button>
        <button className="camera">Toggle Camera</button>
      </div>
    </div>
  );
}

export default Chat;
