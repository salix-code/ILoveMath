export {};

const SESSION_KEY = 'ilovmath_session_id';

interface NextResponse {
  guid: string;
  type_label: string;
  content: string;
}

let currentGUID = '';

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

  typeLabel.textContent = data.type_label;
  // Each sentence separated by ；or 。on its own paragraph
  content.innerHTML = data.content
    .replace(/；/g, '；<br>')
    .replace(/。/g, '。<br>');
  answerInput.value = '';
  answerInput.focus();

  currentGUID = data.guid;
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
