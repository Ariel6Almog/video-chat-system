import React, { useState } from "react";
import "./Register.css";
import { useNavigate } from "react-router-dom";

function Register() {
  const navigate = useNavigate();

  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");

  const handleRegister = (e) => {
    e.preventDefault();

    if (!email || !password || !name) {
      setError("All of the fields need to be filled");
      return;
    }

    fetch("http://localhost:8080/register", {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify({ name, email, password }),
})
  .then((res) => {
    if (!res.ok) throw new Error("Registration failed");
    return res.json();
  })
  .then((data) => {
    console.log("User registered:", data);
    navigate("/login");
  })
  .catch((err) => setError(err.message));

    setError("");
    navigate("/login");
  };

  return (
    <div className="register-container">
      <h2>Register Page</h2>
      <form onSubmit={handleRegister}>
        <input
          type="text"
          placeholder="Name"
          value={name}
          onChange={(e) => setName(e.target.value)}
        />

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

        <button type="submit">Register</button>
      </form>
    </div>
  );
}

export default Register;
