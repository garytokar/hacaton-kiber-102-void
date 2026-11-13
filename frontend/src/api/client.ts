import {
  AuthResponse, User, Fragment, Translation, Achievement, DailyQuest,
  WeeklyExpedition, LeaderboardUser, TeamLeaderboardEntry, SchoolLeaderboardEntry,
  DictionaryEntry, TaskSubmissionResult, PublicStats, ActivityPoint, SchoolOption,
  AdminReviewItem
} from '../types';

const BASE_URL = '/api';

function getAuthHeader(): Record<string, string> {
  const token = localStorage.getItem('token');
  return token ? { Authorization: `Bearer ${token}` } : {};
}

async function request<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
  const headers = {
    'Content-Type': 'application/json',
    ...getAuthHeader(),
    ...options.headers,
  };

  const response = await fetch(`${BASE_URL}${endpoint}`, {
    ...options,
    headers,
  });

  if (!response.ok) {
    let errorDetail = 'Произошла ошибка при выполнении запроса';
    try {
      const errorData = await response.json();
      errorDetail = errorData.detail || errorData.message || errorDetail;
    } catch {
      errorDetail = response.statusText || errorDetail;
    }
    throw new Error(errorDetail);
  }

  return response.json();
}

export const api = {
  // Auth
  register: (data: any) => request<AuthResponse>('/auth/register', { method: 'POST', body: JSON.stringify(data) }),
  login: (data: any) => request<AuthResponse>('/auth/login', { method: 'POST', body: JSON.stringify(data) }),
  getMe: () => request<User>('/auth/me'),

  // User
  updateProfile: (data: any) => request<User>('/users/me', { method: 'PATCH', body: JSON.stringify(data) }),
  getUserActivity: () => request<ActivityPoint[]>('/users/me/activity'),
  protectStreak: () => request<{ success: boolean; message: string; currency: number }>('/users/me/protect-streak', { method: 'POST' }),
  getSchools: () => request<SchoolOption[]>('/users/schools'),

  // Fragments & Daily Task
  getDailyFragment: () => request<Fragment>('/fragments/daily'),
  getFragments: (params?: { category?: string; difficulty?: string }) => {
    const q = new URLSearchParams(params as any).toString();
    return request<Fragment[]>(`/fragments${q ? `?${q}` : ''}`);
  },
  getFragmentById: (id: number) => request<Fragment>(`/fragments/${id}`),

  // Translations
  proposeTranslation: (data: { fragment_id: number; translation_text: string }) =>
    request<TaskSubmissionResult>('/translations', { method: 'POST', body: JSON.stringify(data) }),
  voteTranslation: (id: number, vote_type: 'confirm' | 'reject') =>
    request<TaskSubmissionResult>(`/translations/${id}/vote`, { method: 'POST', body: JSON.stringify({ vote_type }) }),
  getTranslationHistory: (id: number) => request<any[]>(`/translations/${id}/history`),

  // Quests & Expeditions
  getDailyQuests: () => request<DailyQuest[]>('/quests/daily'),
  claimQuest: (id: number) => request<TaskSubmissionResult>(`/quests/${id}/claim`, { method: 'POST' }),
  getWeeklyExpedition: () => request<WeeklyExpedition>('/quests/expedition'),

  // Leaderboard
  getLeaderboard: (period = 'all_time', type = 'xp') =>
    request<LeaderboardUser[]>(`/leaderboard?period=${period}&type=${type}`),
  getTeamLeaderboard: () => request<TeamLeaderboardEntry[]>('/leaderboard/teams'),
  getSchoolLeaderboard: () => request<SchoolLeaderboardEntry[]>('/leaderboard/schools'),

  // Achievements
  getAchievements: () => request<Achievement[]>('/achievements'),

  // Dictionary
  getDictionary: (params?: { search?: string; favorite_only?: boolean; difficulty?: string }) => {
    const searchParams = new URLSearchParams();
    if (params?.search) searchParams.append('search', params.search);
    if (params?.favorite_only) searchParams.append('favorite_only', 'true');
    if (params?.difficulty) searchParams.append('difficulty', params.difficulty);
    const q = searchParams.toString();
    return request<DictionaryEntry[]>(`/dictionary${q ? `?${q}` : ''}`);
  },
  addDictionaryWord: (data: { fragment_id?: number; word: string; translation: string; context?: string; difficulty?: string }) =>
    request<DictionaryEntry>('/dictionary', { method: 'POST', body: JSON.stringify(data) }),
  toggleFavoriteWord: (id: number) => request<DictionaryEntry>(`/dictionary/${id}/favorite`, { method: 'PATCH' }),
  deleteDictionaryWord: (id: number) => request<{ success: boolean }>(`/dictionary/${id}`, { method: 'DELETE' }),

  // Stats & Treasury
  getPublicStats: () => request<PublicStats>('/stats/overview'),
  getTreasury: (params?: { search?: string; category?: string }) => {
    const searchParams = new URLSearchParams();
    if (params?.search) searchParams.append('search', params.search);
    if (params?.category) searchParams.append('category', params.category);
    const q = searchParams.toString();
    return request<{ categories: string[]; items: any[] }>(`/stats/treasury${q ? `?${q}` : ''}`);
  },

  // Admin
  getAdminUsers: (params?: { search?: string; role?: string }) => {
    const q = new URLSearchParams(params as any).toString();
    return request<User[]>(`/admin/users${q ? `?${q}` : ''}`);
  },
  updateUserRole: (id: number, role: string) => request<User>(`/admin/users/${id}/role?role=${role}`, { method: 'PATCH' }),
  getReviewQueue: (status?: string) => request<AdminReviewItem[]>(`/admin/review${status ? `?status=${status}` : ''}`),
  approveTranslation: (id: number) => request<{ success: boolean; message: string }>(`/admin/review/${id}/approve`, { method: 'POST' }),
  rejectTranslation: (id: number, reason?: string) =>
    request<{ success: boolean; message: string }>(`/admin/review/${id}/reject${reason ? `?reason=${encodeURIComponent(reason)}` : ''}`, { method: 'POST' }),
  addFragmentAdmin: (data: any) => request<{ success: boolean; id: number; message: string }>('/admin/fragments', { method: 'POST', body: JSON.stringify(data) }),
  getAdminStats: () => request<any>('/admin/stats'),
};
