"use strict";
(() => {
  // src/question.ts
  var SESSION_KEY = "ilovmath_session_id";
  var currentGUID = "";
  var timerInterval = null;
  var startTime = 0;
  function startTimer() {
    const timerEl = document.getElementById("q-timer");
    if (timerInterval) clearInterval(timerInterval);
    startTime = Date.now();
    timerInterval = window.setInterval(() => {
      const elapsed = Math.floor((Date.now() - startTime) / 1e3);
      const mm = Math.floor(elapsed / 60).toString().padStart(2, "0");
      const ss = (elapsed % 60).toString().padStart(2, "0");
      timerEl.textContent = `${mm}:${ss}`;
    }, 1e3);
  }
  async function fetchNext(prevGUID, prevAnswer) {
    const sessionId = localStorage.getItem(SESSION_KEY) ?? "";
    const res = await fetch("/api/question/next", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "X-Session-ID": sessionId
      },
      body: JSON.stringify({ prev_guid: prevGUID, prev_answer: prevAnswer })
    });
    if (!res.ok) {
      throw new Error(`API error: ${res.status}`);
    }
    return res.json();
  }
  function renderQuestion(data) {
    const typeLabel = document.getElementById("q-type-label");
    const content = document.getElementById("q-content");
    const answerInput = document.getElementById("q-answer");
    const scoreEl = document.getElementById("q-score");
    const totalEl = document.getElementById("q-total");
    typeLabel.textContent = data.type_label;
    scoreEl.textContent = data.score.toString();
    totalEl.textContent = data.total.toString();
    content.innerHTML = data.content.replace(/；/g, "\uFF1B<br>").replace(/。/g, "\u3002<br>");
    answerInput.value = "";
    answerInput.focus();
    currentGUID = data.guid;
    startTimer();
  }
  async function load(prevGUID = "", prevAnswer = "") {
    const container = document.getElementById("board");
    try {
      const data = await fetchNext(prevGUID, prevAnswer);
      renderQuestion(data);
      container.classList.remove("board--error");
    } catch (err) {
      document.getElementById("q-content").textContent = "\u52A0\u8F7D\u9898\u76EE\u5931\u8D25\uFF0C\u8BF7\u8FD4\u56DE\u91CD\u8BD5\u3002";
      console.error(err);
    }
  }
  function onSubmit() {
    const answer = document.getElementById("q-answer").value.trim();
    load(currentGUID, answer);
  }
  function onSolve() {
    window.open("/solution?guid=" + currentGUID, "_blank");
  }
  document.addEventListener("DOMContentLoaded", () => {
    document.getElementById("btn-submit").addEventListener("click", onSubmit);
    document.getElementById("btn-solve").addEventListener("click", onSolve);
    document.getElementById("q-answer").addEventListener("keydown", (e) => {
      if (e.key === "Enter") onSubmit();
    });
    load();
  });
})();
