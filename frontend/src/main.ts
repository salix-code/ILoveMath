const SESSION_KEY = 'ilovmath_session_id';
const ORDER_KEY = 'ilovmath_order';

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
      <div class="difficulty-toggle">
        <button class="diff-btn selected" data-difficulty="1">低</button>
        <button class="diff-btn" data-difficulty="2">中</button>
        <button class="diff-btn" data-difficulty="3">高</button>
      </div>
      <div class="card-actions">
        <button class="btn btn-primary" data-id="${p.id}" data-action="start">开始</button>
        <button class="btn btn-secondary" data-id="${p.id}" data-action="print">打印</button>
      </div>
    `;
    grid.appendChild(card);
  }

  // Delegate click handling for difficulty toggle and action buttons
  grid.addEventListener('click', async (e) => {
    const target = e.target as HTMLElement;

    // Difficulty capsule: toggle selection within the same card
    const diffBtn = target.closest<HTMLButtonElement>('.diff-btn');
    if (diffBtn) {
      const toggle = diffBtn.closest('.difficulty-toggle')!;
      toggle.querySelectorAll('.diff-btn').forEach(b => b.classList.remove('selected'));
      diffBtn.classList.add('selected');
      return;
    }

    const btn = target.closest<HTMLButtonElement>('button[data-action]');
    if (!btn) return;
    const id = Number(btn.dataset['id']);
    const action = btn.dataset['action'];

    if (action === 'start' || action === 'print')  {
      const card = btn.closest('.card')!;
      const selected = card.querySelector<HTMLButtonElement>('.diff-btn.selected');
      const difficulty = Number(selected?.dataset['difficulty'] ?? 1);
      const sessionId = localStorage.getItem(SESSION_KEY) ?? '';
      const res = await fetch('/api/question/start', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'X-Session-ID': sessionId,
        },
        body: JSON.stringify({ id, difficulty, action, order: localStorage.getItem(ORDER_KEY) ?? 'random' }),
      });
      if (res.ok) {
        const data = await res.json() as { redirect: string };
        window.location.href = data.redirect;
      } else {
        console.error('[ILoveMath] start failed', res.status);
      }
    } else {
      console.log(`[ILoveMath] action=${action} problem_id=${id}`);
    }
  });
}

function initSettings(): void {
  const overlay = document.getElementById('settings-overlay')!;
  const openBtn = document.getElementById('settings-btn')!;
  const closeBtn = document.getElementById('settings-close')!;
  const radios = document.querySelectorAll<HTMLInputElement>('input[name="order"]');

  // Restore saved preference (default: random)
  const saved = localStorage.getItem(ORDER_KEY) ?? 'random';
  for (const r of radios) {
    r.checked = r.value === saved;
  }

  openBtn.addEventListener('click', () => overlay.classList.remove('hidden'));
  closeBtn.addEventListener('click', () => overlay.classList.add('hidden'));
  overlay.addEventListener('click', (e) => {
    if (e.target === overlay) overlay.classList.add('hidden');
  });

  for (const r of radios) {
    r.addEventListener('change', () => {
      if (r.checked) localStorage.setItem(ORDER_KEY, r.value);
    });
  }
}

async function init(): Promise<void> {
  initSettings();
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
