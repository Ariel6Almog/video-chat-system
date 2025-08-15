import React, { useState } from "react";
import "./Login.css";
import { useNavigate, Link } from "react-router-dom";

function Login() {
  const navigate = useNavigate();

  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");

  const handleLogin = (e) => {
    e.preventDefault();
    

    if (!email || !password) {
      setError("All of the fields need to be filled");
      return;
    }

    //Send request to the server for validation
    //////////////////////////
    
    fetch("http://localhost:8080/auth/api/login", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ email, password }),
    })
      .then((res) => {
        if (!res.ok) throw new Error("Login failed");
        return res.json();
      })
      .then((data) => {
        console.log("User logged in:", data);
        // שמור טוקן אם יש
        // localStorage.setItem("token", data.token);
        navigate("/chat");
      })
      .catch((err) => setError(err.message));

    setError("");
    navigate("/chat");
  };

  return (
    <div className="login-container">
      <h2>Login Page</h2>
      <form onSubmit={handleLogin}>
        <input
          type="email"
          placeholder="Email"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
        />

        <input
          type="password"
          placeholder="Password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
        />

        {error && <p className="error">{error}</p>}

        <button type="submit">Login</button>
      </form>
      <Link to="/register">Create an Account</Link>
    </div>
  );
}

export default Login;
