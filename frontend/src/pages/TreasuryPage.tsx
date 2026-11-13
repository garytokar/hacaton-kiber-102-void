import React, { useEffect, useState } from 'react';
import { api } from '../api/client';
import { useAuth } from '../context/AuthContext';
import treasurySunrise from '../assets/figma/treasury_sunrise.png';
import {
  Gem, Search, Filter, BookOpen, Volume2, Bookmark, CheckCircle2,
  Download, Sparkles, Copy, Check, Heart, ArrowRight
} from 'lucide-react';

export const TreasuryPage: React.FC = () => {
  const { user } = useAuth();
  const [items, setItems] = useState<any[]>([]);
  const [categories, setCategories] = useState<string[]>(['Все']);
  const [selectedCat, setSelectedCat] = useState('Все');
  const [search, setSearch] = useState('');
  const [loading, setLoading] = useState(true);
  const [copiedId, setCopiedId] = useState<number | null>(null);

  const loadTreasury = async () => {
    setLoading(true);
    try {
      const data = await api.getTreasury({
        category: selectedCat !== 'Все' ? selectedCat : undefined,
        search: search.trim() || undefined,
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

  const handleSearchSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    loadTreasury();
  };

  const handleCopy = (id: number, text: string) => {
    navigator.clipboard.writeText(text);
    setCopiedId(id);
    setTimeout(() => setCopiedId(null), 2000);
  };

  const handleAddToDict = async (item: any) => {
    if (!user) {
      alert('Войдите в аккаунт, чтобы сохранять слова в словарь');
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
      alert('Слово добавлено в ваш словарь!');
    } catch (err: any) {
      alert(err.message || 'Ошибка добавления');
    }
  };

  return (
    <div className="space-y-6 max-w-6xl mx-auto pb-16 select-none">
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
          className="px-5 py-2.5 rounded-full bg-white hover:bg-slate-100 text-[#0C2340] font-bold text-xs shadow-md flex items-center gap-2 transition-all shrink-0"
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
              <span className="w-1 h-4 bg-[#1E50A0] rounded-full inline-block" />
              <span className="text-xs font-bold text-slate-400 uppercase tracking-wider">
                Жемчужина сегодняшнего дня
              </span>
            </div>
            <div className="text-3xl sm:text-4xl font-black text-[#0C2340] tracking-tight">
              «ҡояш» — солнце
            </div>
            <p className="text-sm text-slate-600 leading-relaxed max-w-xl">
              Одно из древнейших сакральных слов тюркских языков. В башкирской традиции солнце олицетворяет жизнь, свет знаний и тепло домашнего очага.
            </p>
            <div className="flex flex-wrap items-center gap-2 pt-1">
              <span className="px-3 py-1 rounded-full bg-[#EEF2FF] text-[#1E50A0] text-xs font-bold">
                ba-ru
              </span>
              <span className="px-3 py-1 rounded-full bg-emerald-50 text-emerald-700 text-xs font-bold flex items-center gap-1">
                <Heart className="w-3.5 h-3.5 fill-emerald-600" />
                <span>128 подтверждений</span>
              </span>
              <span className="px-3 py-1 rounded-full bg-slate-100 text-slate-700 text-xs font-semibold">
                Автор: Айгуль Х. (Башкирская гимназия №1)
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
            className="w-full pl-10 pr-28 py-2.5 rounded-full bg-[#F1F5F9] border-none focus:outline-none focus:ring-2 focus:ring-[#1E50A0] text-xs sm:text-sm text-slate-900"
          />
          <button
            type="submit"
            className="absolute right-1.5 top-1.5 px-4 py-1.5 rounded-full bg-[#0C2340] text-white text-xs font-bold hover:bg-[#1A4480] transition-colors"
          >
            Найти
          </button>
        </form>

        {/* Category Pills */}
        <div className="flex items-center gap-1.5 overflow-x-auto pb-1 pt-1">
          {categories.map((cat) => (
            <button
              key={cat}
              onClick={() => setSelectedCat(cat)}
              className={`px-3.5 py-1.5 rounded-full text-xs font-bold whitespace-nowrap transition-colors ${
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
        <div className="bg-white rounded-3xl p-12 text-center border border-slate-200 text-slate-400 text-sm">
          Записей по вашему запросу не найдено
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
                  <span className="text-[10px] font-bold px-2.5 py-0.5 rounded-full bg-[#EEF2FF] text-[#1E50A0]">
                    {item.category}
                  </span>
                  <div className="flex items-center gap-1 text-[11px] font-bold text-emerald-700 bg-emerald-50 px-2 py-0.5 rounded-full">
                    <CheckCircle2 className="w-3.5 h-3.5" />
                    <span>+{item.votes_count}</span>
                  </div>
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
                <span className="truncate max-w-[150px] font-medium">{item.source_title}</span>
                <div className="flex items-center gap-1">
                  <button
                    onClick={() => handleCopy(item.id, `${item.text_ba} — ${item.translation_text}`)}
                    className="p-1.5 rounded-lg hover:bg-slate-100 text-slate-500 transition-colors"
                    title="Копировать"
                  >
                    {copiedId === item.id ? (
                      <Check className="w-3.5 h-3.5 text-emerald-600" />
                    ) : (
                      <Copy className="w-3.5 h-3.5" />
                    )}
                  </button>
                  <button
                    onClick={() => handleAddToDict(item)}
                    className="p-1.5 rounded-lg hover:bg-blue-50 text-slate-500 hover:text-[#1E50A0] transition-colors"
                    title="Добавить в словарь"
                  >
                    <Bookmark className="w-3.5 h-3.5" />
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
