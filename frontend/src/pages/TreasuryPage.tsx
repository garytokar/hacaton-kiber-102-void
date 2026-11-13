import React, { useEffect, useState } from 'react';
import { api } from '../api/client';
import { useAuth } from '../context/AuthContext';
import treasurySunrise from '../assets/figma/treasury_sunrise.png';
import {
  Gem, Search, Filter, BookOpen, Volume2, Bookmark, CheckCircle2,
  Download, Sparkles, Copy, Check, Heart, ArrowRight, X, AlertCircle
} from 'lucide-react';
import { AudioSpeakerButton } from '../components/ui/AudioSpeakerButton';

const BASHKIR_LETTERS = ['Ә', 'Ө', 'Ү', 'Ҡ', 'Ғ', 'Ң', 'Ҙ', 'Ҫ', 'Һ'];

export const TreasuryPage: React.FC = () => {
  const { user } = useAuth();
  const [items, setItems] = useState<any[]>([]);
  const [categories, setCategories] = useState<string[]>(['Все']);
  const [selectedCat, setSelectedCat] = useState('Все');
  const [search, setSearch] = useState('');
  const [loading, setLoading] = useState(true);
  const [copiedId, setCopiedId] = useState<number | null>(null);
  const [savedIds, setSavedIds] = useState<number[]>([]);
  const [toast, setToast] = useState<{ message: string; type: 'success' | 'info' | 'error' } | null>(null);

  const showToast = (message: string, type: 'success' | 'info' | 'error' = 'success') => {
    setToast({ message, type });
    setTimeout(() => setToast(null), 3000);
  };

  const loadTreasury = async (customSearch?: string) => {
    setLoading(true);
    try {
      const q = customSearch !== undefined ? customSearch : search;
      const data = await api.getTreasury({
        category: selectedCat !== 'Все' ? selectedCat : undefined,
        search: q.trim() || undefined,
      });
      setItems(data.items);
      setCategories(data.categories);
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadTreasury();
  }, [selectedCat]);

  // Debounced search on input change
  useEffect(() => {
    const timer = setTimeout(() => {
      loadTreasury(search);
    }, 300);
    return () => clearTimeout(timer);
  }, [search]);

  const handleSearchSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    loadTreasury();
  };

  const handleCopy = (id: number, text: string) => {
    navigator.clipboard.writeText(text);
    setCopiedId(id);
    showToast('Текст скопирован в буфер обмена! 📋', 'success');
    setTimeout(() => setCopiedId(null), 2000);
  };

  const handleAddToDict = async (item: any) => {
    if (!user) {
      showToast('Войдите в аккаунт, чтобы сохранять слова в персональный словарь', 'info');
      return;
    }
    if (savedIds.includes(item.id)) {
      showToast('Этот фрагмент уже добавлен в ваш словарь!', 'info');
      return;
    }
    try {
      await api.addDictionaryWord({
        fragment_id: item.fragment_id,
        word: item.text_ba,
        translation: item.translation_text,
        context: item.source_title,
        difficulty: item.difficulty,
      });
      setSavedIds((prev) => [...prev, item.id]);
      showToast('Фрагмент успешно добавлен в ваш словарь! 📖', 'success');
    } catch (err: any) {
      showToast(err.message || 'Ошибка добавления', 'error');
    }
  };

  return (
    <div className="space-y-6 max-w-6xl mx-auto pb-16 relative">
      {/* Toast Notification */}
      {toast && (
        <div className="fixed bottom-6 left-1/2 -translate-x-1/2 z-50 animate-fade-in shadow-2xl rounded-2xl px-5 py-3 flex items-center gap-2.5 text-xs font-bold bg-[#0C2340] text-white border border-slate-700">
          {toast.type === 'success' && <CheckCircle2 className="w-4 h-4 text-emerald-400 shrink-0" />}
          {toast.type === 'error' && <AlertCircle className="w-4 h-4 text-red-400 shrink-0" />}
          {toast.type === 'info' && <Sparkles className="w-4 h-4 text-blue-400 shrink-0" />}
          <span>{toast.message}</span>
        </div>
      )}

      {/* Header Banner */}
      <div className="bg-[#0C2340] rounded-3xl p-6 sm:p-8 text-white shadow-md flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4 border border-slate-700/50">
        <div>
          <div className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full bg-[#1E50A0]/60 text-xs font-semibold backdrop-blur mb-2 border border-blue-400/30">
            <Gem className="w-3.5 h-3.5 text-amber-400" />
            <span>Цифровое наследие народа</span>
          </div>
          <h1 className="text-2xl sm:text-3xl font-extrabold tracking-tight">
            Общая сокровищница языка
          </h1>
          <p className="text-slate-300 text-xs sm:text-sm mt-1 max-w-xl">
            База проверенных и подтверждённых сообществом переводов башкирского языка, фольклора и эпоса «Урал-батыр».
          </p>
        </div>

        <a
          href="/api/export/translations?format=csv"
          download
          className="px-5 py-2.5 rounded-full bg-white hover:bg-slate-100 text-[#0C2340] font-bold text-xs shadow-md flex items-center gap-2 transition-all shrink-0 cursor-pointer"
        >
          <Download className="w-4 h-4 text-[#1E50A0]" />
          <span>Экспорт корпуса (CSV)</span>
        </a>
      </div>

      {/* Featured Word of the Treasury (Matching Figma Sunrise Card) */}
      <div className="bg-white rounded-3xl border border-slate-200/90 p-6 shadow-sm">
        <div className="grid grid-cols-1 md:grid-cols-12 gap-6 items-center">
          <div className="md:col-span-8 space-y-3">
            <div className="flex items-center gap-2">
              <span className="px-3 py-1 rounded-full bg-amber-50 text-amber-800 text-xs font-bold border border-amber-200/60 flex items-center gap-1">
                <Sparkles className="w-3.5 h-3.5 text-amber-500" />
                <span>Жемчужина сегодняшнего дня</span>
              </span>
              <span className="text-xs text-slate-400 font-medium">128 подтверждений</span>
            </div>

            <div className="flex items-center gap-3">
              <h2 className="text-3xl sm:text-4xl font-black text-[#0C2340] font-serif">
                ҡояш
              </h2>
              <AudioSpeakerButton text="ҡояш" size="md" />
            </div>

            <p className="text-sm sm:text-base font-semibold text-slate-700">
              солнце • небесное светило, символ жизни и тепла в башкирской поэзии
            </p>

            <div className="text-xs text-slate-500 italic bg-[#F8FAFC] p-3 rounded-xl border border-slate-100">
              «Ҡояш сыҡты — донъя яҡтырҙы» (Взошло солнце — озарился мир)
            </div>

            <div className="flex items-center gap-3 pt-1">
              <span className="px-3 py-1 rounded-full bg-[#EEF2FF] text-[#1E50A0] text-xs font-bold">
                Категория: Природа и космос
              </span>
              <span className="px-3 py-1 rounded-full bg-slate-100 text-slate-700 text-xs font-semibold">
                Автор: Айгуль Х. (БРГИ №1)
              </span>
            </div>
          </div>

          <div className="md:col-span-4 rounded-2xl overflow-hidden border border-slate-100 shadow-inner">
            <img
              src={treasurySunrise}
              alt="Восход над горами"
              className="w-full h-36 object-cover hover:scale-105 transition-transform duration-300"
            />
          </div>
        </div>
      </div>

      {/* Filter and Search Bar */}
      <div className="bg-white rounded-2xl p-4 border border-slate-200/90 shadow-sm space-y-3">
        <form onSubmit={handleSearchSubmit} className="relative">
          <Search className="w-4 h-4 text-slate-400 absolute left-3.5 top-3.5" />
          <input
            type="text"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Поиск по башкирскому тексту, переводу или источнику..."
            className="w-full pl-10 pr-24 py-2.5 rounded-full bg-[#F1F5F9] border-none focus:outline-none focus:ring-2 focus:ring-[#1E50A0] text-xs sm:text-sm text-slate-900"
          />
          {search && (
            <button
              type="button"
              onClick={() => setSearch('')}
              className="absolute right-20 top-3 text-slate-400 hover:text-slate-600"
              title="Очистить"
            >
              <X className="w-4 h-4" />
            </button>
          )}
          <button
            type="submit"
            className="absolute right-1.5 top-1.5 px-4 py-1.5 rounded-full bg-[#0C2340] text-white text-xs font-bold hover:bg-[#1A4480] transition-colors cursor-pointer"
          >
            Найти
          </button>
        </form>

        {/* Bashkir Quick Letters Pill Bar */}
        <div className="flex items-center gap-1.5 overflow-x-auto pb-1">
          <span className="text-[11px] font-bold text-slate-400 shrink-0">Башкирские буквы:</span>
          {BASHKIR_LETTERS.map((letter) => (
            <button
              key={letter}
              type="button"
              onClick={() => setSearch((prev) => prev + letter.toLowerCase())}
              className="w-7 h-7 rounded-lg bg-slate-50 hover:bg-blue-50 border border-slate-200 hover:border-blue-400 text-xs font-bold text-[#0C2340] shrink-0 transition-all cursor-pointer shadow-2xs"
              title={`Вставить букву ${letter}`}
            >
              {letter}
            </button>
          ))}
        </div>

        {/* Category Pills */}
        <div className="flex items-center gap-1.5 overflow-x-auto pb-1 pt-1 border-t border-slate-100">
          {categories.map((cat) => (
            <button
              key={cat}
              onClick={() => setSelectedCat(cat)}
              className={`px-3.5 py-1.5 rounded-full text-xs font-bold whitespace-nowrap transition-colors cursor-pointer ${
                selectedCat === cat
                  ? 'bg-[#1E50A0] text-white shadow-sm'
                  : 'bg-slate-100 text-slate-600 hover:bg-slate-200'
              }`}
            >
              {cat}
            </button>
          ))}
        </div>
      </div>

      {/* Cards Grid */}
      {loading ? (
        <div className="text-center py-12 text-slate-400 text-sm">
          Загрузка сокровищницы...
        </div>
      ) : items.length === 0 ? (
        <div className="bg-white rounded-3xl p-12 text-center border border-slate-200 text-slate-400 text-sm space-y-3">
          <Gem className="w-10 h-10 mx-auto text-slate-300" />
          <p className="font-semibold text-slate-600">Записей по вашему запросу не найдено</p>
          <p className="text-xs text-slate-400">Попробуйте изменить поисковый запрос или выбрать другую категорию</p>
          <button
            onClick={() => {
              setSearch('');
              setSelectedCat('Все');
            }}
            className="px-4 py-2 rounded-full bg-[#1E50A0] text-white text-xs font-bold hover:bg-blue-700 transition-colors cursor-pointer"
          >
            Сбросить фильтры
          </button>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {items.map((item) => (
            <div
              key={item.id}
              className="bg-white rounded-3xl p-5 border border-slate-200/90 shadow-sm hover:shadow-md transition-shadow flex flex-col justify-between space-y-4"
            >
              <div className="space-y-3">
                <div className="flex items-center justify-between">
                  <span className="text-[10px] font-bold px-2.5 py-1 rounded-full bg-slate-100 text-slate-600 uppercase">
                    {item.category || 'Архив'}
                  </span>
                  <span className="text-xs font-bold text-emerald-600 flex items-center gap-1">
                    <Check className="w-3.5 h-3.5" />
                    <span>+{item.votes_count || 1}</span>
                  </span>
                </div>

                <div>
                  <div className="text-base sm:text-lg font-extrabold text-[#0C2340] leading-snug font-serif">
                    «{item.text_ba}»
                  </div>
                  <div className="text-xs sm:text-sm font-semibold text-slate-700 mt-2">
                    {item.translation_text}
                  </div>
                </div>
              </div>

              <div className="pt-3 border-t border-slate-100 flex items-center justify-between text-xs text-slate-400">
                <span className="truncate max-w-[150px] font-medium" title={item.source_title}>
                  {item.source_title}
                </span>
                <div className="flex items-center gap-1.5">
                  <AudioSpeakerButton text={item.text_ba} size="sm" />
                  <button
                    onClick={() => handleCopy(item.id, `${item.text_ba} — ${item.translation_text}`)}
                    className="p-1.5 rounded-lg hover:bg-slate-100 text-slate-500 transition-colors cursor-pointer"
                    title="Копировать цитату"
                  >
                    {copiedId === item.id ? (
                      <Check className="w-3.5 h-3.5 text-emerald-600" />
                    ) : (
                      <Copy className="w-3.5 h-3.5" />
                    )}
                  </button>
                  <button
                    onClick={() => handleAddToDict(item)}
                    className={`p-1.5 rounded-lg transition-colors cursor-pointer ${
                      savedIds.includes(item.id)
                        ? 'bg-blue-50 text-[#1E50A0]'
                        : 'hover:bg-blue-50 text-slate-500 hover:text-[#1E50A0]'
                    }`}
                    title={savedIds.includes(item.id) ? 'Уже в словаре' : 'Сохранить в словарь'}
                  >
                    <Bookmark className={`w-3.5 h-3.5 ${savedIds.includes(item.id) ? 'fill-[#1E50A0]' : ''}`} />
                  </button>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};
