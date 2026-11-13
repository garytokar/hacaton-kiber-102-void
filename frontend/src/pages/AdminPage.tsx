import React, { useEffect, useState } from 'react';
import { api } from '../api/client';
import { useAuth } from '../context/AuthContext';
import { AdminReviewItem, User } from '../types';
import { AudioSpeakerButton } from '../components/ui/AudioSpeakerButton';
import {
  ShieldCheck, Users, CheckCircle2, XCircle, Plus, Download,
  BarChart2, RefreshCw, AlertTriangle, BookOpen, Trash2, School,
  Sparkles, Check, ChevronRight, Layers, FileText
} from 'lucide-react';

interface AdminFragmentItem {
  id: number;
  text_ba: string;
  source_title: string;
  category: string;
  difficulty: string;
  translations_count: number;
  confirmed_count: number;
}

export const AdminPage: React.FC = () => {
  const { user } = useAuth();
  const [activeTab, setActiveTab] = useState<'review' | 'fragments' | 'add_fragment' | 'users' | 'schools' | 'stats'>('review');
  const [reviewItems, setReviewItems] = useState<AdminReviewItem[]>([]);
  const [fragmentsList, setFragmentsList] = useState<AdminFragmentItem[]>([]);
  const [usersList, setUsersList] = useState<User[]>([]);
  const [schoolsList, setSchoolsList] = useState<any[]>([]);
  const [stats, setStats] = useState<any>(null);
  const [loading, setLoading] = useState(true);
  const [feedbackNotice, setFeedbackNotice] = useState<string | null>(null);

  // New Fragment form state
  const [newTextBa, setNewTextBa] = useState('');
  const [newSourceTitle, setNewSourceTitle] = useState('');
  const [newSourceDoc, setNewSourceDoc] = useState('Корпус башкирского языка');
  const [newCategory, setNewCategory] = useState('Пословицы и поговорки');
  const [newDifficulty, setNewDifficulty] = useState('medium');
  const [newTranslation, setNewTranslation] = useState('');

  // New School form state
  const [newSchoolName, setNewSchoolName] = useState('');
  const [newSchoolCity, setNewSchoolCity] = useState('Уфа');
  const [newSchoolClasses, setNewSchoolClasses] = useState('5-А, 6-А, 7-А, 8-А, 9-А, 10-А, 11-А');

  const showNotice = (msg: string) => {
    setFeedbackNotice(msg);
    setTimeout(() => setFeedbackNotice(null), 4000);
  };

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

  const loadFragments = async () => {
    setLoading(true);
    try {
      const data = await api.getAdminFragments();
      setFragmentsList(data);
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

  const loadSchools = async () => {
    setLoading(true);
    try {
      const data = await api.getSchools();
      setSchoolsList(data);
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
    else if (activeTab === 'fragments') loadFragments();
    else if (activeTab === 'users') loadUsers();
    else if (activeTab === 'schools') loadSchools();
    else if (activeTab === 'stats') loadStats();
  }, [activeTab]);

  const handleApprove = async (id: number) => {
    try {
      const res = await api.approveTranslation(id);
      showNotice(res.message);
      loadReviewQueue();
    } catch (err: any) {
      alert(err.message || 'Ошибка одобрения');
    }
  };

  const handleReject = async (id: number) => {
    const reason = prompt('Укажите причину отклонения перевода:');
    if (reason === null) return;
    try {
      const res = await api.rejectTranslation(id, reason);
      showNotice(res.message);
      loadReviewQueue();
    } catch (err: any) {
      alert(err.message || 'Ошибка отклонения');
    }
  };

  const handleRoleChange = async (userId: number, newRole: string) => {
    try {
      await api.updateUserRole(userId, newRole);
      showNotice(`Роль пользователя успешно изменена на ${newRole}`);
      loadUsers();
    } catch (err: any) {
      alert(err.message || 'Ошибка смены роли');
    }
  };

  const handleDeleteFragment = async (id: number) => {
    if (!confirm('Вы уверены, что хотите удалить этот фрагмент и его переводы?')) return;
    try {
      const res = await api.deleteAdminFragment(id);
      showNotice(res.message);
      loadFragments();
    } catch (err: any) {
      alert(err.message || 'Ошибка удаления');
    }
  };

  const handleApplyPreset = (preset: 'oral_batyr' | 'mustay' | 'proverb') => {
    if (preset === 'oral_batyr') {
      setNewTextBa('Күккә ҡарап ай күрҙем, айҙың артында май күрҙем.');
      setNewSourceTitle('Эпос «Урал-батыр»');
      setNewSourceDoc('Башкирский героический эпос, песнь 3');
      setNewCategory('Эпос');
      setNewDifficulty('hard');
      setNewTranslation('Взглянул на небо — увидел месяц, а за месяцем — изобилие и благодать.');
    } else if (preset === 'mustay') {
      setNewTextBa('Беҙҙең өйҙөң йәме — яҡшы кешеләрҙә.');
      setNewSourceTitle('Мустай Карим');
      setNewSourceDoc('Повесть «Радость нашего дома»');
      setNewCategory('Литература');
      setNewDifficulty('medium');
      setNewTranslation('Красота и тепло нашего дома — в добрых и искренних людях.');
    } else if (preset === 'proverb') {
      setNewTextBa('Яҡшы һүҙ — йән аҙығы, яман һүҙ — баш ҡаҙығы.');
      setNewSourceTitle('Башкирские народные пословицы');
      setNewSourceDoc('Сборник фольклора АН РБ');
      setNewCategory('Пословицы и поговорки');
      setNewDifficulty('easy');
      setNewTranslation('Доброе слово — услада для души, а злое слово — заноза для головы.');
    }
  };

  const handleCreateFragment = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      await api.addFragmentAdmin({
        text_ba: newTextBa,
        source_title: newSourceTitle,
        source_document: newSourceDoc,
        category: newCategory,
        difficulty: newDifficulty,
        initial_translation: newTranslation,
      });
      showNotice('Фрагмент успешно добавлен в обучающую базу заданий!');
      setNewTextBa('');
      setNewSourceTitle('');
      setNewTranslation('');
      setActiveTab('fragments');
    } catch (err: any) {
      alert(err.message || 'Ошибка добавления');
    }
  };

  const handleCreateSchool = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newSchoolName.trim()) return;

    try {
      const classNames = newSchoolClasses.split(',').map((c) => c.trim()).filter(Boolean);
      const res = await api.createAdminSchool({
        name: newSchoolName.trim(),
        city: newSchoolCity.trim() || 'Уфа',
        class_names: classNames,
      });
      showNotice(res.message);
      setNewSchoolName('');
      loadSchools();
    } catch (err: any) {
      alert(err.message || 'Ошибка создания школы');
    }
  };

  return (
    <div className="space-y-6 max-w-6xl mx-auto pb-16">
      {/* Banner */}
      <div className="bg-[#0C2340] rounded-3xl p-6 sm:p-8 text-white shadow-md flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4 border border-slate-700/50">
        <div>
          <div className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full bg-[#1E50A0]/60 text-xs font-semibold backdrop-blur mb-2 border border-blue-400/30">
            <ShieldCheck className="w-3.5 h-3.5 text-amber-400" />
            <span>Панель экспертной модерации и управления</span>
          </div>
          <h1 className="text-2xl sm:text-3xl font-extrabold tracking-tight">
            Управление корпусом и модерация ЙӘНЛӘНДЕР
          </h1>
          <p className="text-slate-300 text-xs sm:text-sm mt-1">
            Проверка спорных переводов, пополнение базы текстов, управление школами и экспорт датасетов.
          </p>
        </div>

        <div className="flex items-center gap-2">
          <a
            href="/api/export/translations?format=csv"
            download
            className="px-4 py-2.5 rounded-full bg-white hover:bg-slate-100 text-[#0C2340] font-bold text-xs shadow-md flex items-center gap-2 transition-all shrink-0"
          >
            <Download className="w-4 h-4 text-[#1E50A0]" />
            <span>CSV датасет</span>
          </a>
          <a
            href="/api/export/translations?format=json"
            download
            className="px-4 py-2.5 rounded-full bg-[#1E50A0] hover:bg-[#2563EB] text-white font-bold text-xs shadow-md flex items-center gap-2 transition-all shrink-0"
          >
            <Download className="w-4 h-4" />
            <span>JSON датасет</span>
          </a>
        </div>
      </div>

      {feedbackNotice && (
        <div className="bg-emerald-50 border border-emerald-200 text-emerald-800 text-xs p-4 rounded-2xl flex items-center gap-2 shadow-xs">
          <CheckCircle2 className="w-4 h-4 text-emerald-600 shrink-0" />
          <span className="font-semibold">{feedbackNotice}</span>
        </div>
      )}

      {/* Tabs */}
      <div className="flex items-center gap-1.5 bg-white p-2 rounded-2xl border border-slate-200/90 shadow-sm overflow-x-auto">
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
          onClick={() => setActiveTab('fragments')}
          className={`px-4 py-2 rounded-xl text-xs font-bold whitespace-nowrap transition-all ${
            activeTab === 'fragments'
              ? 'bg-[#1E50A0] text-white shadow-sm'
              : 'text-slate-600 hover:bg-slate-100'
          }`}
        >
          Корпус текстов ({fragmentsList.length || '...'})
        </button>
        <button
          onClick={() => setActiveTab('add_fragment')}
          className={`px-4 py-2 rounded-xl text-xs font-bold whitespace-nowrap transition-all ${
            activeTab === 'add_fragment'
              ? 'bg-[#1E50A0] text-white shadow-sm'
              : 'text-slate-600 hover:bg-slate-100'
          }`}
        >
          + Добавить материал
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
          onClick={() => setActiveTab('schools')}
          className={`px-4 py-2 rounded-xl text-xs font-bold whitespace-nowrap transition-all ${
            activeTab === 'schools'
              ? 'bg-[#1E50A0] text-white shadow-sm'
              : 'text-slate-600 hover:bg-slate-100'
          }`}
        >
          Школы и классы
        </button>
        <button
          onClick={() => setActiveTab('stats')}
          className={`px-4 py-2 rounded-xl text-xs font-bold whitespace-nowrap transition-all ${
            activeTab === 'stats'
              ? 'bg-[#1E50A0] text-white shadow-sm'
              : 'text-slate-600 hover:bg-slate-100'
          }`}
        >
          Статистика
        </button>
      </div>

      {/* Tab: Review Queue */}
      {activeTab === 'review' && (
        <div className="bg-white rounded-3xl border border-slate-200/90 shadow-sm overflow-hidden">
          {loading ? (
            <div className="text-center py-12 text-slate-400 text-sm">Загрузка очереди...</div>
          ) : reviewItems.length === 0 ? (
            <div className="p-12 text-center text-slate-400 text-sm">
              Очередь модерации чиста! Все предложенные варианты проверены экспертами.
            </div>
          ) : (
            <div className="divide-y divide-slate-100">
              {reviewItems.map((item) => (
                <div key={item.id} className="p-5 sm:p-6 space-y-3 hover:bg-slate-50/50 transition-colors">
                  <div className="flex flex-wrap items-center justify-between gap-2">
                    <span className="text-[10px] font-bold px-2.5 py-0.5 rounded-full bg-amber-100 text-amber-800 uppercase">
                      {item.status === 'expert_review' ? 'Спорный перевод (Экспертная проверка)' : 'Новое предложение'}
                    </span>
                    <span className="text-xs text-slate-400">
                      Автор: <strong className="text-slate-700">@{item.creator_username}</strong> • {new Date(item.created_at).toLocaleDateString('ru-RU')}
                    </span>
                  </div>

                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4 pt-1">
                    <div className="p-4 rounded-2xl bg-[#F8FAFC] border border-slate-200/70 space-y-2">
                      <div className="flex items-center justify-between">
                        <span className="text-[10px] font-bold text-slate-400 uppercase">Оригинал на башкирском</span>
                        <AudioSpeakerButton text={item.text_ba} size="sm" />
                      </div>
                      <div className="text-base font-black text-[#0C2340] font-serif leading-snug">
                        «{item.text_ba}»
                      </div>
                      <div className="text-xs text-slate-500 font-semibold">
                        Источник: {item.source_title}
                      </div>
                    </div>

                    <div className="p-4 rounded-2xl bg-blue-50/50 border border-blue-100 space-y-2">
                      <span className="text-[10px] font-bold text-[#1E50A0] uppercase">Предложенный перевод</span>
                      <div className="text-base font-bold text-slate-900 leading-snug">
                        «{item.translation_text}»
                      </div>
                      <div className="flex items-center gap-3 text-xs text-slate-500 pt-1">
                        <span>Голосов за: <strong className="text-emerald-700 font-bold">+{item.votes_count}</strong></span>
                        <span>Голосов против: <strong className="text-red-600 font-bold">-{item.reject_count}</strong></span>
                      </div>
                    </div>
                  </div>

                  <div className="flex items-center justify-end gap-2 pt-2">
                    <button
                      onClick={() => handleReject(item.translation_id)}
                      className="px-4 py-2 rounded-full border border-red-200 text-red-600 hover:bg-red-50 text-xs font-bold flex items-center gap-1.5 transition-colors"
                    >
                      <XCircle className="w-4 h-4" />
                      <span>Отклонить</span>
                    </button>
                    <button
                      onClick={() => handleApprove(item.translation_id)}
                      className="px-5 py-2 rounded-full bg-emerald-600 hover:bg-emerald-700 text-white text-xs font-bold flex items-center gap-1.5 shadow-sm transition-colors"
                    >
                      <CheckCircle2 className="w-4 h-4" />
                      <span>Одобрить и в Сокровищницу</span>
                    </button>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      )}

      {/* Tab: Fragments Corpus List */}
      {activeTab === 'fragments' && (
        <div className="bg-white rounded-3xl border border-slate-200/90 shadow-sm overflow-hidden">
          <div className="p-5 border-b border-slate-100 flex items-center justify-between">
            <div>
              <h3 className="font-extrabold text-base text-[#0C2340]">Языковой корпус текстов</h3>
              <p className="text-xs text-slate-500">Всего фрагментов в учебной базе: {fragmentsList.length}</p>
            </div>
            <button
              onClick={() => setActiveTab('add_fragment')}
              className="px-4 py-2 rounded-full bg-[#1E50A0] hover:bg-[#2563EB] text-white text-xs font-bold flex items-center gap-1.5 shadow-sm transition-all"
            >
              <Plus className="w-3.5 h-3.5" />
              <span>Добавить материал</span>
            </button>
          </div>

          {loading ? (
            <div className="text-center py-12 text-slate-400 text-sm">Загрузка корпуса...</div>
          ) : (
            <div className="divide-y divide-slate-100">
              {fragmentsList.map((frag) => (
                <div key={frag.id} className="p-4 sm:p-5 flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4 hover:bg-slate-50 transition-colors">
                  <div className="space-y-1.5 flex-1">
                    <div className="flex items-center gap-2">
                      <span className="text-[10px] font-bold px-2 py-0.5 rounded-full bg-[#EEF2FF] text-[#1E50A0]">
                        {frag.category}
                      </span>
                      <span className="text-[10px] font-bold px-2 py-0.5 rounded-full bg-slate-100 text-slate-600">
                        {frag.difficulty === 'hard' ? 'Сложный' : frag.difficulty === 'easy' ? 'Лёгкий' : 'Средний'}
                      </span>
                      <AudioSpeakerButton text={frag.text_ba} size="sm" />
                    </div>
                    <div className="text-sm sm:text-base font-black text-[#0C2340] font-serif">
                      «{frag.text_ba}»
                    </div>
                    <div className="text-xs text-slate-400">
                      Источник: {frag.source_title} • Вариантов перевода: {frag.translations_count} (Подтверждено: {frag.confirmed_count})
                    </div>
                  </div>

                  <div className="flex items-center gap-2 self-end sm:self-center">
                    <button
                      onClick={() => handleDeleteFragment(frag.id)}
                      className="p-2 rounded-xl text-slate-300 hover:text-red-500 hover:bg-red-50 transition-colors"
                      title="Удалить фрагмент"
                    >
                      <Trash2 className="w-4 h-4" />
                    </button>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      )}

      {/* Tab: Add Fragment Form with Presets */}
      {activeTab === 'add_fragment' && (
        <div className="bg-white rounded-3xl border border-slate-200/90 shadow-sm p-6 sm:p-8 max-w-3xl mx-auto space-y-6">
          <div className="flex items-center justify-between">
            <div>
              <h3 className="text-xl font-black text-[#0C2340]">
                Добавить материал в корпус башкирского языка
              </h3>
              <p className="text-xs text-slate-500 mt-0.5">
                Загрузите новый отрывок литературного произведения, пословицу или статью.
              </p>
            </div>
            <FileText className="w-8 h-8 text-[#1E50A0]/40" />
          </div>

          {/* Quick Presets for Demo */}
          <div className="p-4 rounded-2xl bg-amber-50/70 border border-amber-200/70 space-y-2">
            <div className="text-xs font-bold text-amber-900 flex items-center gap-1.5">
              <Sparkles className="w-4 h-4 text-amber-600" />
              <span>Быстрые шаблоны для автозаполнения формы:</span>
            </div>
            <div className="flex flex-wrap gap-2 text-xs">
              <button
                type="button"
                onClick={() => handleApplyPreset('oral_batyr')}
                className="px-3 py-1.5 rounded-full bg-white hover:bg-amber-100 border border-amber-300 text-amber-900 font-bold transition-colors"
              >
                Песнь из «Урал-батыр»
              </button>
              <button
                type="button"
                onClick={() => handleApplyPreset('mustay')}
                className="px-3 py-1.5 rounded-full bg-white hover:bg-amber-100 border border-amber-300 text-amber-900 font-bold transition-colors"
              >
                Мустай Карим
              </button>
              <button
                type="button"
                onClick={() => handleApplyPreset('proverb')}
                className="px-3 py-1.5 rounded-full bg-white hover:bg-amber-100 border border-amber-300 text-amber-900 font-bold transition-colors"
              >
                Народная пословица
              </button>
            </div>
          </div>

          <form onSubmit={handleCreateFragment} className="space-y-4">
            <div>
              <div className="flex items-center justify-between mb-1">
                <label className="block text-xs font-bold text-slate-700">
                  Фрагмент текста на башкирском языке *
                </label>
                {newTextBa && <AudioSpeakerButton text={newTextBa} size="sm" showLabel={true} label="Прослушать текст" />}
              </div>
              <textarea
                required
                rows={3}
                value={newTextBa}
                onChange={(e) => setNewTextBa(e.target.value)}
                placeholder="Мәҫәлән: Тыуған тел — халыҡтың рухи байлығы."
                className="w-full p-4 rounded-2xl bg-slate-50 border border-slate-200 text-sm text-slate-900 focus:bg-white focus:outline-none focus:ring-2 focus:ring-[#1E50A0] transition-all font-serif"
              />
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <div>
                <label className="block text-xs font-bold text-slate-700 mb-1">
                  Источник / Произведение / Автор *
                </label>
                <input
                  type="text"
                  required
                  value={newSourceTitle}
                  onChange={(e) => setNewSourceTitle(e.target.value)}
                  placeholder="Например: Эпос «Урал-батыр» или Зайнаб Биишева"
                  className="w-full px-4 py-2.5 rounded-full bg-slate-50 border border-slate-200 text-xs sm:text-sm focus:bg-white focus:outline-none focus:ring-2 focus:ring-[#1E50A0]"
                />
              </div>

              <div>
                <label className="block text-xs font-bold text-slate-700 mb-1">
                  Контекст или документ
                </label>
                <input
                  type="text"
                  value={newSourceDoc}
                  onChange={(e) => setNewSourceDoc(e.target.value)}
                  placeholder="Том, глава, страница или архивный источник"
                  className="w-full px-4 py-2.5 rounded-full bg-slate-50 border border-slate-200 text-xs sm:text-sm focus:bg-white focus:outline-none focus:ring-2 focus:ring-[#1E50A0]"
                />
              </div>
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-xs font-bold text-slate-700 mb-1">Категория</label>
                <select
                  value={newCategory}
                  onChange={(e) => setNewCategory(e.target.value)}
                  className="w-full px-4 py-2.5 rounded-full bg-slate-50 border border-slate-200 text-xs sm:text-sm focus:bg-white focus:outline-none focus:ring-2 focus:ring-[#1E50A0]"
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
                  className="w-full px-4 py-2.5 rounded-full bg-slate-50 border border-slate-200 text-xs sm:text-sm focus:bg-white focus:outline-none focus:ring-2 focus:ring-[#1E50A0]"
                >
                  <option value="easy">Лёгкий (+5 XP)</option>
                  <option value="medium">Средний (+5 XP)</option>
                  <option value="hard">Сложный (+8 XP)</option>
                </select>
              </div>
            </div>

            <div>
              <label className="block text-xs font-bold text-slate-700 mb-1">
                Эталонный или черновой перевод на русский язык
              </label>
              <input
                type="text"
                value={newTranslation}
                onChange={(e) => setNewTranslation(e.target.value)}
                placeholder="Перевод для старта голосования сообщества"
                className="w-full px-4 py-2.5 rounded-full bg-slate-50 border border-slate-200 text-xs sm:text-sm focus:bg-white focus:outline-none focus:ring-2 focus:ring-[#1E50A0]"
              />
            </div>

            <button
              type="submit"
              className="w-full py-3.5 rounded-full bg-[#0C2340] hover:bg-[#1A4480] text-white font-bold text-xs sm:text-sm shadow-md transition-all cursor-pointer"
            >
              Сохранить и опубликовать в корпус заданий
            </button>
          </form>
        </div>
      )}

      {/* Tab: Users Management */}
      {activeTab === 'users' && (
        <div className="bg-white rounded-3xl border border-slate-200/90 shadow-sm overflow-hidden">
          <div className="p-5 border-b border-slate-100">
            <h3 className="font-extrabold text-base text-[#0C2340]">Пользователи платформы</h3>
            <p className="text-xs text-slate-500">Назначение ролей: ученик, учитель, эксперт-лингвист или администратор.</p>
          </div>

          <div className="divide-y divide-slate-100">
            {usersList.map((u) => (
              <div key={u.id} className="p-4 sm:p-5 flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4 hover:bg-slate-50 transition-colors">
                <div className="space-y-1">
                  <div className="flex items-center gap-2">
                    <span className="font-bold text-slate-900 text-sm">{u.full_name || u.username}</span>
                    <span className="text-xs text-slate-400">@{u.username}</span>
                    <span className="text-[10px] font-bold px-2 py-0.5 rounded-full bg-[#EEF2FF] text-[#1E50A0] uppercase">
                      {u.role}
                    </span>
                  </div>
                  <div className="text-xs text-slate-500">
                    {u.email} • Школа: {u.school_name || 'Не указана'}
                  </div>
                </div>

                <div className="flex items-center gap-2">
                  <span className="text-xs text-slate-400 font-semibold">Сменить роль:</span>
                  <select
                    value={u.role}
                    onChange={(e) => handleRoleChange(u.id, e.target.value)}
                    className="px-3 py-1.5 rounded-full border border-slate-200 text-xs font-bold text-slate-700 bg-white focus:outline-none focus:ring-2 focus:ring-[#1E50A0]"
                  >
                    <option value="student">Ученик (student)</option>
                    <option value="teacher">Учитель (teacher)</option>
                    <option value="expert">Эксперт (expert)</option>
                    <option value="admin">Администратор (admin)</option>
                  </select>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Tab: Schools & Classes Management */}
      {activeTab === 'schools' && (
        <div className="space-y-6">
          {/* Add School Form */}
          <div className="bg-white rounded-3xl border border-slate-200/90 shadow-sm p-6 sm:p-8 space-y-4">
            <div className="flex items-center justify-between">
              <div>
                <h3 className="text-lg font-black text-[#0C2340]">Добавить учебное заведение</h3>
                <p className="text-xs text-slate-500">Подключение школ, лицеев, гимназий и классов к общему рейтингу.</p>
              </div>
              <School className="w-6 h-6 text-[#1E50A0]" />
            </div>

            <form onSubmit={handleCreateSchool} className="space-y-3">
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                <div>
                  <label className="block text-xs font-bold text-slate-700 mb-1">Название школы *</label>
                  <input
                    type="text"
                    required
                    value={newSchoolName}
                    onChange={(e) => setNewSchoolName(e.target.value)}
                    placeholder="Например: Лицей №153 г. Уфы"
                    className="w-full px-4 py-2 rounded-full border border-slate-200 text-xs sm:text-sm focus:outline-none focus:ring-2 focus:ring-[#1E50A0]"
                  />
                </div>
                <div>
                  <label className="block text-xs font-bold text-slate-700 mb-1">Город / Район *</label>
                  <input
                    type="text"
                    required
                    value={newSchoolCity}
                    onChange={(e) => setNewSchoolCity(e.target.value)}
                    placeholder="Уфа, Стерлитамак, Сибай..."
                    className="w-full px-4 py-2 rounded-full border border-slate-200 text-xs sm:text-sm focus:outline-none focus:ring-2 focus:ring-[#1E50A0]"
                  />
                </div>
              </div>

              <div>
                <label className="block text-xs font-bold text-slate-700 mb-1">
                  Классы / Группы (через запятую)
                </label>
                <input
                  type="text"
                  value={newSchoolClasses}
                  onChange={(e) => setNewSchoolClasses(e.target.value)}
                  placeholder="5-А, 6-А, 7-А, 8-А, 9-А"
                  className="w-full px-4 py-2 rounded-full border border-slate-200 text-xs sm:text-sm focus:outline-none focus:ring-2 focus:ring-[#1E50A0]"
                />
              </div>

              <button
                type="submit"
                className="py-2.5 px-6 rounded-full bg-[#0C2340] hover:bg-[#1A4480] text-white font-bold text-xs shadow-md transition-all"
              >
                + Зарегистрировать школу
              </button>
            </form>
          </div>

          {/* Existing Schools List */}
          <div className="bg-white rounded-3xl border border-slate-200/90 shadow-sm p-6 space-y-4">
            <h4 className="font-extrabold text-sm text-[#0C2340]">Подключённые учебные заведения:</h4>
            <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-3">
              {schoolsList.map((s: any) => (
                <div key={s.id} className="p-4 rounded-2xl bg-[#F8FAFC] border border-slate-200/70 space-y-2">
                  <div className="font-bold text-slate-800 text-sm">{s.name}</div>
                  <div className="text-xs text-slate-400">г. {s.city}</div>
                  <div className="flex flex-wrap gap-1 pt-1">
                    {s.classes?.map((c: any) => (
                      <span key={c.id} className="px-2 py-0.5 rounded-md bg-white border border-slate-200 text-[10px] text-slate-600 font-semibold">
                        {c.name}
                      </span>
                    ))}
                  </div>
                </div>
              ))}
            </div>
          </div>
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
