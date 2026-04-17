const SESSION_KEY = 'ilovmath_session_id';

interface QuestionItem {
  id: number;
  title: string;
}

interface ListResponse {
  session_id: string;
  items: QuestionItem[];
}

async function fetchQuestions(): Promise<QuestionItem[]> {
  const sessionId = localStorage.getItem(SESSION_KEY);
  const headers: HeadersInit = { 'Content-Type': 'application/json' };
  if (sessionId) headers['X-Session-ID'] = sessionId;

  const res = await fetch('/api/list', { headers });
  if (!res.ok) throw new Error(`${res.status}`);

  const data = (await res.json()) as ListResponse;
  if (data.session_id) localStorage.setItem(SESSION_KEY, data.session_id);
  return data.items;
}

function renderList(items: QuestionItem[]): void {
  const ol = document.querySelector<HTMLOListElement>('.q-list')!;
  ol.innerHTML = '';
  for (let i = 0; i < items.length; i++) {
    const li = document.createElement('li');
    li.className = 'q-item';
    li.innerHTML = `
      <div class="q-body">
        <span class="q-num">${i + 1}.</span>
        <span class="q-text">${items[i].title}</span>
      </div>
      <div class="answer">答：<span class="answer-line"></span></div>`;
    ol.appendChild(li);
  }
}

function renderError(): void {
  const ol = document.querySelector<HTMLOListElement>('.q-list')!;
  ol.innerHTML = '<li class="q-item">无法获取题目列表，请刷新重试。</li>';
}

document.addEventListener('DOMContentLoaded', async () => {
  try {
    const items = await fetchQuestions();
    items.length > 0 ? renderList(items) : renderError();
  } catch {
    renderError();
  }
});