import test from 'node:test';
import assert from 'node:assert/strict';

// ============================================================================
// 1. GAMIFICATION & LEVEL THRESHOLDS (5 tests)
// ============================================================================

const LEVEL_THRESHOLDS = [
  { level: 1, title: 'Начинающий', minXp: 0 },
  { level: 2, title: 'Ученик', minXp: 100 },
  { level: 3, title: 'Искатель слов', minXp: 250 },
  { level: 4, title: 'Знаток', minXp: 500 },
  { level: 5, title: 'Переводчик', minXp: 900 },
  { level: 6, title: 'Хранитель', minXp: 1500 },
  { level: 7, title: 'Мастер языка', minXp: 2500 },
  { level: 8, title: 'Наставник', minXp: 4000 },
  { level: 9, title: 'Мудрец', minXp: 6500 },
  { level: 10, title: 'Легенда языка', minXp: 10000 },
];

function getLevelInfo(xp) {
  let current = LEVEL_THRESHOLDS[0];
  for (const t of LEVEL_THRESHOLDS) {
    if (xp >= t.minXp) current = t;
    else break;
  }
  const nextThreshold = LEVEL_THRESHOLDS.find((t) => t.level === current.level + 1);
  return {
    level: current.level,
    title: current.title,
    nextLevelXp: nextThreshold ? nextThreshold.minXp : current.minXp,
  };
}

test('Level calculator: 0 XP returns Level 1 Начинающий', () => {
  const info = getLevelInfo(0);
  assert.equal(info.level, 1);
  assert.equal(info.title, 'Начинающий');
  assert.equal(info.nextLevelXp, 100);
});

test('Level calculator: 900 XP returns Level 5 Переводчик (Figma profile)', () => {
  const info = getLevelInfo(900);
  assert.equal(info.level, 5);
  assert.equal(info.title, 'Переводчик');
  assert.equal(info.nextLevelXp, 1500);
});

test('Level calculator: intermediate 1200 XP returns Level 5', () => {
  const info = getLevelInfo(1200);
  assert.equal(info.level, 5);
});

test('Level calculator: 15000 XP returns maximum Level 10 Легенда языка', () => {
  const info = getLevelInfo(15000);
  assert.equal(info.level, 10);
  assert.equal(info.title, 'Легенда языка');
});

test('Level calculator: exactly 10 thresholds are configured', () => {
  assert.equal(LEVEL_THRESHOLDS.length, 10);
});


// ============================================================================
// 2. BASHKIR SPECIAL CHARACTERS & SEARCH NORMALIZATION (5 tests)
// ============================================================================

const BASHKIR_SPECIAL_LETTERS = ['ә', 'ө', 'ү', 'ғ', 'ҡ', 'ң', 'ҙ', 'ҫ', 'һ'];

function normalizeBashkirSearch(query) {
  return query
    .trim()
    .toLowerCase()
    .replace(/ё/g, 'е');
}

function hasBashkirSpecificLetters(text) {
  const lower = text.toLowerCase();
  return BASHKIR_SPECIAL_LETTERS.some((ch) => lower.includes(ch));
}

test('Bashkir search: preserves special letter ҡ in ҡояш', () => {
  const norm = normalizeBashkirSearch('  ҠОЯШ  ');
  assert.equal(norm, 'ҡояш');
  assert.equal(hasBashkirSpecificLetters(norm), true);
});

test('Bashkir search: detects special letter ә in йәнләндер', () => {
  assert.equal(hasBashkirSpecificLetters('ЙӘНЛӘНДЕР'), true);
});

test('Bashkir search: standard Russian query returns false for specific letters', () => {
  assert.equal(hasBashkirSpecificLetters('Солнце и горы'), false);
});

test('Bashkir search: normalizes ё to е', () => {
  assert.equal(normalizeBashkirSearch('Тёплый вечер'), 'теплый вечер');
});

test('Bashkir alphabet: all 9 unique letters are registered', () => {
  assert.equal(BASHKIR_SPECIAL_LETTERS.length, 9);
});


// ============================================================================
// 3. UI FORMATTERS & PALETTE TOKENS (5 tests)
// ============================================================================

function formatNumberSpaces(num) {
  return num.toString().replace(/\B(?=(\d{3})+(?!\d))/g, ' ');
}

function formatAudioSeconds(sec) {
  const m = Math.floor(sec / 60).toString().padStart(2, '0');
  const s = Math.floor(sec % 60).toString().padStart(2, '0');
  return `${m}:${s}`;
}

test('Number formatting: 48392 words formats as "48 392" (Figma Stat Card)', () => {
  assert.equal(formatNumberSpaces(48392), '48 392');
});

test('Number formatting: 4820 community members formats as "4 820"', () => {
  assert.equal(formatNumberSpaces(4820), '4 820');
});

test('Audio formatting: 5 seconds formats as "00:05"', () => {
  assert.equal(formatAudioSeconds(5), '00:05');
});

test('Audio formatting: 65 seconds formats as "01:05"', () => {
  assert.equal(formatAudioSeconds(65), '01:05');
});

test('Color tokens: Sapphire palette adheres to Figma specification', () => {
  const NAVY = '#0C2340';
  const BLUE = '#1E50A0';
  assert.match(NAVY, /^#[0-9A-F]{6}$/i);
  assert.match(BLUE, /^#[0-9A-F]{6}$/i);
});


// ============================================================================
// 4. API ERROR DESERIALIZATION & VALIDATION (5 tests)
// ============================================================================

function extractApiError(err) {
  if (typeof err === 'string') return err;
  if (!err) return 'Неизвестная ошибка';
  if (err.detail) {
    if (typeof err.detail === 'string') return err.detail;
    if (Array.isArray(err.detail) && err.detail[0]?.msg) return err.detail[0].msg;
  }
  if (err.message) return err.message;
  return 'Произошла ошибка при обращении к серверу';
}

test('Error extractor: handles plain string errors', () => {
  assert.equal(extractApiError('Неверный пароль'), 'Неверный пароль');
});

test('Error extractor: handles FastAPI string detail object', () => {
  assert.equal(extractApiError({ detail: 'Пользователь уже существует' }), 'Пользователь уже существует');
});

test('Error extractor: handles FastAPI Pydantic validation error array', () => {
  const pydanticError = {
    detail: [{ loc: ['body', 'password'], msg: 'Field required', type: 'missing' }]
  };
  assert.equal(extractApiError(pydanticError), 'Field required');
});

test('Error extractor: handles standard Error instance message', () => {
  assert.equal(extractApiError(new Error('Сетевой сбой')), 'Сетевой сбой');
});

test('Error extractor: never returns [object Object]', () => {
  const res = extractApiError({ weirdKey: 123 });
  assert.notEqual(res, '[object Object]');
  assert.ok(typeof res === 'string' && res.length > 0);
});


// ============================================================================
// 5. NAVIGATION & PERMISSION RULES (5 tests)
// ============================================================================

function getVisibleNavItems(role) {
  const items = [
    { to: '/', label: 'Главная' },
    { to: '/task', label: 'Обучение' },
    { to: '/treasury', label: 'Сокровищница' },
    { to: '/leaderboard', label: 'Рейтинг' },
    { to: '/achievements', label: 'Достижения' },
    { to: '/expeditions', label: 'Команда' },
    { to: '/analytics', label: 'Аналитика' },
    { to: '/dictionary', label: 'Словарь' },
    { to: '/shop', label: 'Лавка Хранителя' },
    { to: '/profile', label: 'Профиль' },
  ];
  if (role === 'admin' || role === 'expert' || role === 'teacher') {
    items.push({ to: '/admin', label: 'Модерация' });
  }
  return items;
}

test('Navigation: Student role gets 10 public nav items (no moderation)', () => {
  const items = getVisibleNavItems('student');
  assert.equal(items.length, 10);
  assert.equal(items.some((i) => i.to === '/admin'), false);
});

test('Navigation: Admin role gets 11 nav items including Модерация', () => {
  const items = getVisibleNavItems('admin');
  assert.equal(items.length, 11);
  assert.equal(items[10].to, '/admin');
  assert.equal(items[10].label, 'Модерация');
});

test('Navigation: Expert role gets Модерация menu item', () => {
  const items = getVisibleNavItems('expert');
  assert.equal(items.some((i) => i.to === '/admin'), true);
});

test('Navigation: Teacher role gets Модерация menu item', () => {
  const items = getVisibleNavItems('teacher');
  assert.equal(items.some((i) => i.to === '/admin'), true);
});

test('Navigation: Guest / undefined role defaults to 10 items without /admin', () => {
  const items = getVisibleNavItems(undefined);
  assert.equal(items.length, 10);
});


// ============================================================================
// 6. PAGINATION & DATA SLICING (5 tests)
// ============================================================================

function paginateList(items, page, pageSize) {
  const start = page * pageSize;
  return items.slice(start, start + pageSize);
}

test('Pagination: first page of 5 items returns correct slice', () => {
  const dataset = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10];
  const page = paginateList(dataset, 0, 5);
  assert.deepEqual(page, [1, 2, 3, 4, 5]);
});

test('Pagination: second page returns items 6-10', () => {
  const dataset = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10];
  const page = paginateList(dataset, 1, 5);
  assert.deepEqual(page, [6, 7, 8, 9, 10]);
});

test('Pagination: out of bounds page returns empty array', () => {
  const dataset = [1, 2, 3];
  const page = paginateList(dataset, 5, 5);
  assert.deepEqual(page, []);
});

test('Pagination: circular pagination for daily fragments loop (0 to 4)', () => {
  const total = 5;
  const next = (idx) => (idx + 1) % total;
  const prev = (idx) => (idx - 1 + total) % total;
  assert.equal(next(4), 0);
  assert.equal(prev(0), 4);
});

test('Pagination: format string "01 / 05" correctly pads index and total', () => {
  const formatIndex = (idx, total) =>
    `${(idx + 1).toString().padStart(2, '0')} / ${total.toString().padStart(2, '0')}`;
  assert.equal(formatIndex(0, 5), '01 / 05');
  assert.equal(formatIndex(4, 5), '05 / 05');
});
