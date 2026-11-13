import React, { useEffect, useState } from 'react';
import { api } from '../api/client';
import { useAuth } from '../context/AuthContext';
import { AdminReviewItem, User } from '../types';
import {
  ShieldCheck, Users, CheckCircle2, XCircle, Plus, Download,
  BarChart2, RefreshCw, AlertTriangle
} from 'lucide-react';

export const AdminPage: React.FC = () => {
  const { user } = useAuth();
  const [activeTab, setActiveTab] = useState<'review' | 'users' | 'add_fragment' | 'stats'>('review');
  const [reviewItems, setReviewItems] = useState<AdminReviewItem[]>([]);
  const [usersList, setUsersList] = useState<User[]>([]);
  const [stats, setStats] = useState<any>(null);
  const [loading, setLoading] = useState(true);

  // New Fragment form state
  const [newTextBa, setNewTextBa] = useState('');
  const [newSourceTitle, setNewSourceTitle] = useState('');
  const [newCategory, setNewCategory] = useState('Пословицы и поговорки');
  const [newDifficulty, setNewDifficulty] = useState('medium');
  const [newTranslation, setNewTranslation] = useState('');

  const loadReviewQueue = async () => {
    setLoading(true);
    try {
      const data = await api.getReviewQueue();
      setReviewItems(data);
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  const loadUsers = async () => {
    setLoading(true);
    try {
      const data = await api.getAdminUsers();
      setUsersList(data);
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  const loadStats = async () => {
    setLoading(true);
    try {
      const data = await api.getAdminStats();
      setStats(data);
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (activeTab === 'review') loadReviewQueue();
    else if (activeTab === 'users') loadUsers();
    else if (activeTab === 'stats') loadStats();
  }, [activeTab]);

  const handleApprove = async (id: number) => {
    try {
      await api.approveTranslation(id);
      loadReviewQueue();
    } catch (err: any) {
      alert(err.message || 'Ошибка одобрения');
    }
  };

  const handleReject = async (id: number) => {
    const reason = prompt('Укажите причину отклонения перевода:');
    if (reason === null) return;
    try {
      await api.rejectTranslation(id, reason);
      loadReviewQueue();
    } catch (err: any) {
      alert(err.message || 'Ошибка отклонения');
    }
  };

  const handleRoleChange = async (userId: number, newRole: string) => {
    try {
      await api.updateUserRole(userId, newRole);
      loadUsers();
    } catch (err: any) {
      alert(err.message || 'Ошибка смены роли');
    }
  };

  const handleCreateFragment = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      await api.addFragmentAdmin({
        text_ba: newTextBa,
        source_title: newSourceTitle,
        category: newCategory,
        difficulty: newDifficulty,
        initial_translation: newTranslation,
      });
      alert('Фрагмент успешно добавлен в обучающую базу!');
      setNewTextBa('');
      setNewSourceTitle('');
      setNewTranslation('');
    } catch (err: any) {
      alert(err.message || 'Ошибка добавления');
    }
  };

  return (
    <div className="space-y-6 max-w-6xl mx-auto pb-16 select-none">
      {/* Banner */}
      <div className="bg-[#0C2340] rounded-3xl p-6 sm:p-8 text-white shadow-md flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4 border border-slate-700/50">
        <div>
          <div className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full bg-[#1E50A0]/60 text-xs font-semibold backdrop-blur mb-2 border border-blue-400/30">
            <ShieldCheck className="w-3.5 h-3.5 text-amber-400" />
            <span>Панель экспертной модерации</span>
          </div>
          <h1 className="text-2xl sm:text-3xl font-extrabold tracking-tight">
            Управление и модерация ЙӘНЛӘНДЕР
          </h1>
          <p className="text-slate-300 text-xs sm:text-sm mt-1">
            Проверка спорных переводов, управление корпусом текстов и пользователями.
          </p>
        </div>

        <a
          href="/api/export/translations?format=csv"
          download
          className="px-5 py-2.5 rounded-full bg-white hover:bg-slate-100 text-[#0C2340] font-bold text-xs shadow-md flex items-center gap-2 transition-all shrink-0"
        >
          <Download className="w-4 h-4 text-[#1E50A0]" />
          <span>Экспорт датасета (CSV)</span>
        </a>
      </div>

      {/* Tabs */}
      <div className="flex items-center gap-1.5 bg-white p-2.5 rounded-2xl border border-slate-200/90 shadow-sm overflow-x-auto">
        <button
          onClick={() => setActiveTab('review')}
          className={`px-4 py-2 rounded-xl text-xs font-bold whitespace-nowrap transition-all ${
            activeTab === 'review'
              ? 'bg-[#1E50A0] text-white shadow-sm'
              : 'text-slate-600 hover:bg-slate-100'
          }`}
        >
          Очередь модерации ({reviewItems.length})
        </button>
        <button
          onClick={() => setActiveTab('users')}
          className={`px-4 py-2 rounded-xl text-xs font-bold whitespace-nowrap transition-all ${
            activeTab === 'users'
              ? 'bg-[#1E50A0] text-white shadow-sm'
              : 'text-slate-600 hover:bg-slate-100'
          }`}
        >
          Пользователи
        </button>
        <button
          onClick={() => setActiveTab('add_fragment')}
          className={`px-4 py-2 rounded-xl text-xs font-bold whitespace-nowrap transition-all ${
            activeTab === 'add_fragment'
              ? 'bg-[#1E50A0] text-white shadow-sm'
              : 'text-slate-600 hover:bg-slate-100'
          }`}
        >
          Добавить фрагмент
        </button>
        <button
          onClick={() => setActiveTab('stats')}
          className={`px-4 py-2 rounded-xl text-xs font-bold whitespace-nowrap transition-all ${
            activeTab === 'stats'
              ? 'bg-[#1E50A0] text-white shadow-sm'
              : 'text-slate-600 hover:bg-slate-100'
          }`}
        >
          Статистика платформы
        </button>
      </div>

      {/* Tab: Review Queue */}
      {activeTab === 'review' && (
        <div className="bg-white rounded-3xl border border-slate-200/90 shadow-sm overflow-hidden">
          {loading ? (
            <div className="text-center py-12 text-slate-400 text-sm">Загрузка очереди...</div>
          ) : reviewItems.length === 0 ? (
            <div className="p-12 text-center text-slate-400 text-sm">
              Очередь модерации чиста! Все предложенные варианты проверены.
            </div>
          ) : (
            <div className="divide-y divide-slate-100">
              {reviewItems.map((item) => (
                <div key={item.id} className="p-5 sm:p-6 space-y-3 hover:bg-slate-50/50 transition-colors">
                  <div className="flex flex-wrap items-center justify-between gap-2">
                    <span className="text-[10px] font-bold px-2.5 py-0.5 rounded-full bg-amber-100 text-amber-800 uppercase">
                      {item.status === 'expert_review' ? 'Спорный вариант' : 'На проверке'}
                    </span>
                    <span className="text-xs text-slate-400">
                      Голоса: +{item.votes_count} / -{item.reject_count} • Автор: {item.creator_username || 'Хранитель'}
                    </span>
                  </div>

                  <div>
                    <div className="text-xs text-slate-400 uppercase font-semibold">Оригинал ({item.source_title}):</div>
                    <div className="text-base font-black text-[#0C2340] mt-0.5 font-serif">
                      «{item.text_ba}»
                    </div>
                  </div>

                  <div className="bg-slate-50 p-3.5 rounded-xl border border-slate-200/60">
                    <div className="text-xs text-slate-400 uppercase font-semibold">Предложенный перевод:</div>
                    <div className="text-sm font-bold text-[#1E50A0] mt-0.5">
                      «{item.translation_text}»
                    </div>
                  </div>

                  <div className="flex justify-end gap-2 pt-1">
                    <button
                      onClick={() => handleReject(item.translation_id)}
                      className="px-4 py-2 rounded-full border border-red-200 text-red-600 hover:bg-red-50 text-xs font-bold flex items-center gap-1.5 transition-colors"
                    >
                      <XCircle className="w-3.5 h-3.5" />
                      <span>Отклонить</span>
                    </button>
                    <button
                      onClick={() => handleApprove(item.translation_id)}
                      className="px-5 py-2 rounded-full bg-[#0C2340] hover:bg-[#1A4480] text-white text-xs font-bold flex items-center gap-1.5 shadow transition-colors"
                    >
                      <CheckCircle2 className="w-3.5 h-3.5" />
                      <span>Подтвердить в Сокровищницу</span>
                    </button>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      )}

      {/* Tab: Users Management */}
      {activeTab === 'users' && (
        <div className="bg-white rounded-3xl border border-slate-200/90 shadow-sm overflow-hidden">
          <div className="divide-y divide-slate-100">
            {usersList.map((u) => (
              <div key={u.id} className="p-4 sm:px-6 flex items-center justify-between gap-4 hover:bg-slate-50">
                <div className="flex items-center gap-3">
                  <div className="w-10 h-10 rounded-full bg-[#0C2340] text-white flex items-center justify-center font-bold text-xs uppercase shadow-sm">
                    {u.username.slice(0, 2)}
                  </div>
                  <div>
                    <div className="text-sm font-extrabold text-[#0C2340]">{u.full_name || u.username}</div>
                    <div className="text-xs text-slate-400">{u.email} • {u.school_name || 'Индивидуально'}</div>
                  </div>
                </div>

                <div className="flex items-center gap-3">
                  <span className="text-xs text-slate-500 font-semibold hidden sm:inline">
                    {u.stats?.xp || 0} XP (Ур. {u.stats?.level || 1})
                  </span>
                  <select
                    value={u.role}
                    onChange={(e) => handleRoleChange(u.id, e.target.value)}
                    className="text-xs font-bold px-3 py-1.5 rounded-full border border-slate-300 focus:outline-none focus:ring-2 focus:ring-[#1E50A0] bg-white text-slate-800"
                  >
                    <option value="student">student</option>
                    <option value="teacher">teacher</option>
                    <option value="expert">expert</option>
                    <option value="admin">admin</option>
                  </select>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Tab: Add Fragment */}
      {activeTab === 'add_fragment' && (
        <div className="bg-white rounded-3xl border border-slate-200/90 shadow-sm p-6 sm:p-8 max-w-2xl mx-auto space-y-4">
          <h3 className="text-lg font-black text-[#0C2340]">
            Добавить фрагмент в языковой корпус
          </h3>
          <form onSubmit={handleCreateFragment} className="space-y-4">
            <div>
              <label className="block text-xs font-bold text-slate-700 mb-1">
                Фрагмент текста на башкирском языке *
              </label>
              <textarea
                required
                rows={3}
                value={newTextBa}
                onChange={(e) => setNewTextBa(e.target.value)}
                placeholder="Мәҫәлән: Тыуған тел — халыҡтың рухи байлығы."
                className="w-full p-3.5 rounded-2xl border border-slate-200 text-sm focus:outline-none focus:ring-2 focus:ring-[#1E50A0]"
              />
            </div>

            <div>
              <label className="block text-xs font-bold text-slate-700 mb-1">
                Источник / Произведение *
              </label>
              <input
                type="text"
                required
                value={newSourceTitle}
                onChange={(e) => setNewSourceTitle(e.target.value)}
                placeholder="Например: Эпос «Урал-батыр» или Мустай Карим"
                className="w-full px-4 py-2 rounded-full border border-slate-200 text-sm focus:outline-none focus:ring-2 focus:ring-[#1E50A0]"
              />
            </div>

            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="block text-xs font-bold text-slate-700 mb-1">Категория</label>
                <select
                  value={newCategory}
                  onChange={(e) => setNewCategory(e.target.value)}
                  className="w-full px-4 py-2 rounded-full border border-slate-200 text-xs sm:text-sm focus:outline-none focus:ring-2 focus:ring-[#1E50A0]"
                >
                  <option value="Пословицы и поговорки">Пословицы и поговорки</option>
                  <option value="Эпос">Эпос</option>
                  <option value="Литература">Литература</option>
                  <option value="Культура">Культура</option>
                  <option value="Поэзия">Поэзия</option>
                  <option value="Общие фразы">Общие фразы</option>
                </select>
              </div>

              <div>
                <label className="block text-xs font-bold text-slate-700 mb-1">Сложность</label>
                <select
                  value={newDifficulty}
                  onChange={(e) => setNewDifficulty(e.target.value)}
                  className="w-full px-4 py-2 rounded-full border border-slate-200 text-xs sm:text-sm focus:outline-none focus:ring-2 focus:ring-[#1E50A0]"
                >
                  <option value="easy">Лёгкий (+5 XP)</option>
                  <option value="medium">Средний (+5 XP)</option>
                  <option value="hard">Сложный (+8 XP)</option>
                </select>
              </div>
            </div>

            <div>
              <label className="block text-xs font-bold text-slate-700 mb-1">
                Предварительный перевод на русский (необязательно)
              </label>
              <input
                type="text"
                value={newTranslation}
                onChange={(e) => setNewTranslation(e.target.value)}
                placeholder="Черновой перевод для начала голосования"
                className="w-full px-4 py-2 rounded-full border border-slate-200 text-sm focus:outline-none focus:ring-2 focus:ring-[#1E50A0]"
              />
            </div>

            <button
              type="submit"
              className="w-full py-3 rounded-full bg-[#0C2340] hover:bg-[#1A4480] text-white font-bold text-xs sm:text-sm shadow-md transition-colors"
            >
              Добавить в базу заданий
            </button>
          </form>
        </div>
      )}

      {/* Tab: System Stats */}
      {activeTab === 'stats' && stats && (
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          <div className="bg-white p-6 rounded-3xl border border-slate-200/90 shadow-sm text-center">
            <span className="text-xs font-bold text-slate-400 uppercase">Пользователей</span>
            <div className="text-3xl font-black text-[#0C2340] mt-2">{stats.total_users}</div>
          </div>
          <div className="bg-white p-6 rounded-3xl border border-slate-200/90 shadow-sm text-center">
            <span className="text-xs font-bold text-slate-400 uppercase">Фрагментов в базе</span>
            <div className="text-3xl font-black text-[#0C2340] mt-2">{stats.total_fragments}</div>
          </div>
          <div className="bg-white p-6 rounded-3xl border border-slate-200/90 shadow-sm text-center">
            <span className="text-xs font-bold text-slate-400 uppercase">Подтверждено</span>
            <div className="text-3xl font-black text-emerald-600 mt-2">{stats.confirmed_translations}</div>
          </div>
          <div className="bg-white p-6 rounded-3xl border border-slate-200/90 shadow-sm text-center">
            <span className="text-xs font-bold text-slate-400 uppercase">В очереди модерации</span>
            <div className="text-3xl font-black text-amber-500 mt-2">{stats.in_review_translations}</div>
          </div>
        </div>
      )}
    </div>
  );
};
