export {};

const SESSION_KEY = 'ilovmath_session_id';

interface NextResponse {
  guid: string;
  type_label: string;
  content: string;
  score: number;
  total: number;
  question_count: number;
  correct?: boolean | null;
}

let currentGUID = '';
let timerInterval: number | null = null;
let startTime: number = 0;

function startTimer(): void {
  const timerEl = document.getElementById('q-timer')!;
  if (timerInterval) clearInterval(timerInterval);
  
  startTime = Date.now();
  timerInterval = window.setInterval(() => {
    const elapsed = Math.floor((Date.now() - startTime) / 1000);
    const mm = Math.floor(elapsed / 60).toString().padStart(2, '0');
    const ss = (elapsed % 60).toString().padStart(2, '0');
    timerEl.textContent = `${mm}:${ss}`;
  }, 1000);
}

async function fetchNext(prevGUID: string, prevAnswer: string): Promise<NextResponse> {
  const sessionId = localStorage.getItem(SESSION_KEY) ?? '';
  const res = await fetch('/api/question/next', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'X-Session-ID': sessionId,
    },
    body: JSON.stringify({ prev_guid: prevGUID, prev_answer: prevAnswer }),
  });
  if (!res.ok) {
    throw new Error(`API error: ${res.status}`);
  }
  return res.json() as Promise<NextResponse>;
}

function renderQuestion(data: NextResponse): void {
  const typeLabel = document.getElementById('q-type-label')!;
  const content   = document.getElementById('q-content')!;
  const answerInput = document.getElementById('q-answer') as HTMLInputElement;
  const scoreEl = document.getElementById('q-score')!;
  const totalEl = document.getElementById('q-total')!;

  typeLabel.textContent = data.type_label;
  scoreEl.textContent = data.score.toString();
  totalEl.textContent = data.total.toString();

  // Each sentence separated by ；or 。on its own paragraph
  content.innerHTML = data.content
    .replace(/；/g, '；<br>')
    .replace(/。/g, '。<br>');
  answerInput.value = '';
  answerInput.focus();

  currentGUID = data.guid;
  startTimer();
}

async function load(prevGUID = '', prevAnswer = ''): Promise<void> {
  const container = document.getElementById('board')!;
  try {
    const data = await fetchNext(prevGUID, prevAnswer);
    renderQuestion(data);
    container.classList.remove('board--error');
  } catch (err) {
    document.getElementById('q-content')!.textContent = '加载题目失败，请返回重试。';
    console.error(err);
  }
}

function onSubmit(): void {
  const answer = (document.getElementById('q-answer') as HTMLInputElement).value.trim();
  load(currentGUID, answer);
}

function onSolve(): void {
  // TODO: open solution page
  window.open('/solution?guid=' + currentGUID, '_blank');
}

document.addEventListener('DOMContentLoaded', () => {
  document.getElementById('btn-submit')!.addEventListener('click', onSubmit);
  document.getElementById('btn-solve')!.addEventListener('click', onSolve);

  // Allow pressing Enter in the answer box to submit
  document.getElementById('q-answer')!.addEventListener('keydown', (e) => {
    if (e.key === 'Enter') onSubmit();
  });

  load();
});
