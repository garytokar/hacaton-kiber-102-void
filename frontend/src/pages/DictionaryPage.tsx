import React, { useEffect, useState } from 'react';
import { api } from '../api/client';
import { useAuth } from '../context/AuthContext';
import { DictionaryEntry } from '../types';
import {
  BookA, Search, Star, Plus, Trash2, RotateCw, Volume2,
  ChevronLeft, ChevronRight, CheckCircle2, Sparkles, BookOpen
} from 'lucide-react';

export const DictionaryPage: React.FC = () => {
  const { user } = useAuth();
  const [entries, setEntries] = useState<DictionaryEntry[]>([]);
  const [search, setSearch] = useState('');
  const [favoriteOnly, setFavoriteOnly] = useState(false);
  const [loading, setLoading] = useState(true);
  const [isTrainerOpen, setIsTrainerOpen] = useState(false);
  const [trainerIndex, setTrainerIndex] = useState(0);
  const [isFlipped, setIsFlipped] = useState(false);

  // Add word modal
  const [addModalOpen, setAddModalOpen] = useState(false);
  const [newWord, setNewWord] = useState('');
  const [newTranslation, setNewTranslation] = useState('');
  const [newContext, setNewContext] = useState('');

  const loadDictionary = async () => {
    setLoading(true);
    try {
      const data = await api.getDictionary({
        search: search.trim() || undefined,
        favorite_only: favoriteOnly,
      });
      setEntries(data);
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadDictionary();
  }, [favoriteOnly]);

  const handleToggleFav = async (id: number) => {
    try {
      await api.toggleFavoriteWord(id);
      loadDictionary();
    } catch (err) {
      console.error(err);
    }
  };

  const handleDelete = async (id: number) => {
    if (!confirm('Удалить слово из словаря?')) return;
    try {
      await api.deleteDictionaryWord(id);
      loadDictionary();
    } catch (err) {
      console.error(err);
    }
  };

  const handleAddSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newWord || !newTranslation) return;
    try {
      await api.addDictionaryWord({
        word: newWord.trim(),
        translation: newTranslation.trim(),
        context: newContext.trim() || undefined,
      });
      setAddModalOpen(false);
      setNewWord('');
      setNewTranslation('');
      setNewContext('');
      loadDictionary();
    } catch (err: any) {
      alert(err.message || 'Ошибка добавления слова');
    }
  };

  const handlePlayAudio = (word: string) => {
    if ('speechSynthesis' in window) {
      const utterance = new SpeechSynthesisUtterance(word);
      utterance.lang = 'ru-RU';
      window.speechSynthesis.speak(utterance);
    }
  };

  return (
    <div className="space-y-6 max-w-5xl mx-auto pb-16 select-none">
      {/* Header Banner */}
      <div className="bg-[#0C2340] rounded-3xl p-6 sm:p-8 text-white shadow-md flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4 border border-slate-700/50">
        <div>
          <div className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full bg-[#1E50A0]/60 text-xs font-semibold backdrop-blur mb-2 border border-blue-400/30">
            <BookA className="w-3.5 h-3.5 text-amber-400" />
            <span>Персональный лексикон</span>
          </div>
          <h1 className="text-2xl sm:text-3xl font-extrabold tracking-tight">
            Мой словарь башкирского языка
          </h1>
          <p className="text-slate-300 text-xs sm:text-sm mt-1">
            Сохранённые слова, фразы и режим интерактивного повторения карточек.
          </p>
        </div>

        <div className="flex items-center gap-2">
          {entries.length > 0 && (
            <button
              onClick={() => {
                setTrainerIndex(0);
                setIsFlipped(false);
                setIsTrainerOpen(true);
              }}
              className="px-4 py-2.5 rounded-full bg-[#1E50A0] hover:bg-[#2563EB] text-white font-bold text-xs shadow-md flex items-center gap-2 transition-all shrink-0"
            >
              <RotateCw className="w-4 h-4" />
              <span>Режим карточек</span>
            </button>
          )}

          <button
            onClick={() => setAddModalOpen(true)}
            className="px-4 py-2.5 rounded-full bg-white text-[#0C2340] hover:bg-slate-100 font-bold text-xs shadow-md flex items-center gap-1.5 transition-all shrink-0"
          >
            <Plus className="w-4 h-4" />
            <span>Добавить слово</span>
          </button>
        </div>
      </div>

      {/* Search & Favorites Filter Bar */}
      <div className="bg-white rounded-2xl p-4 border border-slate-200/90 shadow-sm flex flex-col sm:flex-row items-center justify-between gap-3">
        <div className="relative flex-1 w-full">
          <Search className="w-4 h-4 text-slate-400 absolute left-3.5 top-3" />
          <input
            type="text"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && loadDictionary()}
            placeholder="Поиск по словам и переводам..."
            className="w-full pl-10 pr-4 py-2 rounded-full bg-[#F1F5F9] border-none text-xs sm:text-sm text-slate-900 focus:outline-none focus:ring-2 focus:ring-[#1E50A0]"
          />
        </div>

        <button
          onClick={() => setFavoriteOnly(!favoriteOnly)}
          className={`px-4 py-2 rounded-full text-xs font-bold flex items-center gap-2 transition-colors shrink-0 w-full sm:w-auto justify-center ${
            favoriteOnly
              ? 'bg-amber-100 text-amber-800 border border-amber-300'
              : 'bg-slate-100 text-slate-600 hover:bg-slate-200'
          }`}
        >
          <Star className={`w-3.5 h-3.5 ${favoriteOnly ? 'fill-amber-600' : ''}`} />
          <span>Только избранные</span>
        </button>
      </div>

      {/* Dictionary Items List */}
      {loading ? (
        <div className="text-center py-12 text-slate-400 text-sm">
          Загрузка словаря...
        </div>
      ) : entries.length === 0 ? (
        <div className="bg-white rounded-3xl p-12 text-center border border-slate-200 text-slate-400 text-sm space-y-3">
          <BookA className="w-10 h-10 mx-auto text-slate-300" />
          <p>В словаре пока нет слов. Добавьте свои первые слова или сохраняйте их из ежедневных заданий!</p>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {entries.map((item) => (
            <div
              key={item.id}
              className="bg-white rounded-3xl p-5 border border-slate-200/90 shadow-sm hover:shadow-md transition-shadow flex flex-col justify-between space-y-3"
            >
              <div className="space-y-2">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <span className="text-lg font-black text-[#0C2340] font-serif">
                      {item.word}
                    </span>
                    <button
                      onClick={() => handlePlayAudio(item.word)}
                      className="p-1 rounded-full text-slate-400 hover:text-[#1E50A0] transition-colors"
                      title="Озвучить слово"
                    >
                      <Volume2 className="w-4 h-4" />
                    </button>
                  </div>

                  <div className="flex items-center gap-1">
                    <button
                      onClick={() => handleToggleFav(item.id)}
                      className="p-1.5 rounded-lg text-slate-300 hover:text-amber-500 transition-colors"
                    >
                      <Star className={`w-4 h-4 ${item.is_favorite ? 'fill-amber-500 text-amber-500' : ''}`} />
                    </button>
                    <button
                      onClick={() => handleDelete(item.id)}
                      className="p-1.5 rounded-lg text-slate-300 hover:text-red-500 transition-colors"
                    >
                      <Trash2 className="w-4 h-4" />
                    </button>
                  </div>
                </div>

                <div className="text-sm font-semibold text-slate-700">
                  {item.translation}
                </div>

                {item.context && (
                  <div className="text-xs text-slate-400 italic">
                    Контекст: {item.context}
                  </div>
                )}
              </div>

              <div className="pt-2 border-t border-slate-100 flex items-center justify-between text-[11px] text-slate-400">
                <span>Сложность: {item.difficulty === 'hard' ? 'Высокая' : 'Базовая'}</span>
                <span>Добавлено: {new Date(item.created_at).toLocaleDateString('ru-RU')}</span>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Flashcard Trainer Modal */}
      {isTrainerOpen && entries.length > 0 && (
        <div className="fixed inset-0 z-50 bg-black/60 backdrop-blur-xs flex items-center justify-center p-4">
          <div className="bg-white rounded-3xl max-w-md w-full p-6 text-center shadow-2xl border border-slate-200 space-y-6 animate-scale-up">
            <div className="flex items-center justify-between text-xs text-slate-400">
              <span className="font-bold text-[#0C2340]">Карточка {trainerIndex + 1} из {entries.length}</span>
              <button
                onClick={() => setIsTrainerOpen(false)}
                className="font-bold text-slate-400 hover:text-slate-600 text-lg"
              >
                ✕
              </button>
            </div>

            {/* Flashcard Card Body */}
            <div
              onClick={() => setIsFlipped(!isFlipped)}
              className="h-56 rounded-2xl bg-[#F8FAFC] border-2 border-dashed border-[#1E50A0]/40 flex flex-col items-center justify-center p-6 cursor-pointer hover:bg-blue-50/50 transition-colors shadow-inner"
            >
              {!isFlipped ? (
                <div className="space-y-2">
                  <div className="text-2xl sm:text-3xl font-black text-[#0C2340] font-serif">
                    {entries[trainerIndex].word}
                  </div>
                  <div className="text-xs text-slate-400">Нажмите, чтобы увидеть перевод</div>
                </div>
              ) : (
                <div className="space-y-2 animate-fade-in">
                  <div className="text-xl sm:text-2xl font-bold text-[#1E50A0]">
                    {entries[trainerIndex].translation}
                  </div>
                  {entries[trainerIndex].context && (
                    <div className="text-xs text-slate-500 italic max-w-xs">
                      «{entries[trainerIndex].context}»
                    </div>
                  )}
                </div>
              )}
            </div>

            {/* Navigation buttons */}
            <div className="flex items-center justify-between gap-3">
              <button
                onClick={() => {
                  setTrainerIndex((prev) => (prev - 1 + entries.length) % entries.length);
                  setIsFlipped(false);
                }}
                className="p-3 rounded-full border border-slate-200 hover:bg-slate-100 text-slate-700"
              >
                <ChevronLeft className="w-5 h-5" />
              </button>

              <button
                onClick={() => handlePlayAudio(entries[trainerIndex].word)}
                className="px-4 py-2.5 rounded-full bg-slate-100 hover:bg-slate-200 text-slate-700 text-xs font-bold flex items-center gap-1.5"
              >
                <Volume2 className="w-4 h-4" />
                <span>Озвучить</span>
              </button>

              <button
                onClick={() => {
                  setTrainerIndex((prev) => (prev + 1) % entries.length);
                  setIsFlipped(false);
                }}
                className="p-3 rounded-full bg-[#0C2340] hover:bg-[#1A4480] text-white shadow"
              >
                <ChevronRight className="w-5 h-5" />
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Add Word Modal */}
      {addModalOpen && (
        <div className="fixed inset-0 z-50 bg-black/60 backdrop-blur-xs flex items-center justify-center p-4">
          <form
            onSubmit={handleAddSubmit}
            className="bg-white rounded-3xl max-w-md w-full p-6 space-y-4 shadow-2xl border border-slate-200 animate-scale-up"
          >
            <div className="flex items-center justify-between">
              <h3 className="font-bold text-base text-[#0C2340]">Добавить слово в словарь</h3>
              <button
                type="button"
                onClick={() => setAddModalOpen(false)}
                className="text-slate-400 hover:text-slate-600 text-lg font-bold"
              >
                ✕
              </button>
            </div>

            <div className="space-y-3">
              <div>
                <label className="block text-xs font-bold text-slate-700 mb-1">
                  Слово на башкирском языке:
                </label>
                <input
                  required
                  type="text"
                  value={newWord}
                  onChange={(e) => setNewWord(e.target.value)}
                  placeholder="например, дуҫлыҡ"
                  className="w-full px-4 py-2 rounded-xl border border-slate-200 text-sm focus:outline-none focus:ring-2 focus:ring-[#1E50A0]"
                />
              </div>

              <div>
                <label className="block text-xs font-bold text-slate-700 mb-1">
                  Перевод на русский:
                </label>
                <input
                  required
                  type="text"
                  value={newTranslation}
                  onChange={(e) => setNewTranslation(e.target.value)}
                  placeholder="например, дружба"
                  className="w-full px-4 py-2 rounded-xl border border-slate-200 text-sm focus:outline-none focus:ring-2 focus:ring-[#1E50A0]"
                />
              </div>

              <div>
                <label className="block text-xs font-bold text-slate-700 mb-1">
                  Контекст / Фраза (необязательно):
                </label>
                <input
                  type="text"
                  value={newContext}
                  onChange={(e) => setNewContext(e.target.value)}
                  placeholder="например, Дуҫлыҡ — ҙур байлыҡ"
                  className="w-full px-4 py-2 rounded-xl border border-slate-200 text-sm focus:outline-none focus:ring-2 focus:ring-[#1E50A0]"
                />
              </div>
            </div>

            <div className="flex justify-end gap-2 pt-2">
              <button
                type="button"
                onClick={() => setAddModalOpen(false)}
                className="px-4 py-2 rounded-full text-xs font-semibold text-slate-500 hover:bg-slate-100"
              >
                Отмена
              </button>
              <button
                type="submit"
                className="px-5 py-2 rounded-full bg-[#0C2340] hover:bg-[#1A4480] text-white text-xs font-bold shadow"
              >
                Сохранить
              </button>
            </div>
          </form>
        </div>
      )}
    </div>
  );
};
