const SESSION_KEY = 'ilovmath_session_id';

interface ProblemType {
  id: number;
  title: string;
}

interface ListResponse {
  session_id: string;
  items: ProblemType[];
}

async function fetchList(): Promise<ListResponse> {
  const sessionId = localStorage.getItem(SESSION_KEY);
  const headers: HeadersInit = { 'Content-Type': 'application/json' };
  if (sessionId) {
    headers['X-Session-ID'] = sessionId;
  }

  const res = await fetch('/api/list', { headers });
  if (!res.ok) {
    throw new Error(`API error: ${res.status}`);
  }
  return res.json() as Promise<ListResponse>;
}

function renderCards(problems: ProblemType[]): void {
  const grid = document.getElementById('problem-list')!;
  grid.innerHTML = '';

  for (const p of problems) {
    const card = document.createElement('div');
    card.className = 'card';
    card.innerHTML = `
      <h2 class="card-title">${p.title}</h2>
      <div class="card-actions">
        <button class="btn btn-primary" data-id="${p.id}" data-action="practice">练习</button>
        <button class="btn btn-secondary" data-id="${p.id}" data-action="print">打印</button>
      </div>
    `;
    grid.appendChild(card);
  }

  // Delegate click handling for both buttons
  grid.addEventListener('click', (e) => {
    const btn = (e.target as HTMLElement).closest<HTMLButtonElement>('button[data-action]');
    if (!btn) return;
    const id = btn.dataset['id'];
    const action = btn.dataset['action'];
    console.log(`[ILoveMath] action=${action} problem_id=${id}`);
    // TODO: navigate to practice / print page
  });
}

async function init(): Promise<void> {
  try {
    const data = await fetchList();
    localStorage.setItem(SESSION_KEY, data.session_id);
    renderCards(data.items);
  } catch (err) {
    const grid = document.getElementById('problem-list')!;
    grid.innerHTML = '<span class="error">加载失败，请刷新页面重试。</span>';
    console.error(err);
  }
}

init();
