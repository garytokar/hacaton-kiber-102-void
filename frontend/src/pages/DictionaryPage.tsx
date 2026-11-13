import React, { useEffect, useState } from 'react';
import { api } from '../api/client';
import { useAuth } from '../context/AuthContext';
import { DictionaryEntry } from '../types';
import confetti from 'canvas-confetti';
import {
  BookA, Search, Star, Plus, Trash2, RotateCw, Volume2,
  ChevronLeft, ChevronRight, CheckCircle2, Sparkles, BookOpen,
  X, AlertCircle, Check
} from 'lucide-react';
import { AudioSpeakerButton } from '../components/ui/AudioSpeakerButton';

const BASHKIR_LETTERS = ['Ә', 'Ө', 'Ү', 'Ҡ', 'Ғ', 'Ң', 'Ҙ', 'Ҫ', 'Һ'];

export const DictionaryPage: React.FC = () => {
  const { user } = useAuth();
  const [entries, setEntries] = useState<DictionaryEntry[]>([]);
  const [search, setSearch] = useState('');
  const [favoriteOnly, setFavoriteOnly] = useState(false);
  const [loading, setLoading] = useState(true);
  const [isTrainerOpen, setIsTrainerOpen] = useState(false);
  const [trainerIndex, setTrainerIndex] = useState(0);
  const [isFlipped, setIsFlipped] = useState(false);
  const [knownCount, setKnownCount] = useState(0);

  // Add word modal
  const [addModalOpen, setAddModalOpen] = useState(false);
  const [newWord, setNewWord] = useState('');
  const [newTranslation, setNewTranslation] = useState('');
  const [newContext, setNewContext] = useState('');

  // Toast notifications
  const [toast, setToast] = useState<{ message: string; type: 'success' | 'info' | 'error' } | null>(null);

  const showToast = (message: string, type: 'success' | 'info' | 'error' = 'success') => {
    setToast({ message, type });
    setTimeout(() => setToast(null), 3000);
  };

  const loadDictionary = async (customSearch?: string) => {
    setLoading(true);
    try {
      const q = customSearch !== undefined ? customSearch : search;
      const data = await api.getDictionary({
        search: q.trim() || undefined,
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

  // Debounced search on input
  useEffect(() => {
    const timer = setTimeout(() => {
      loadDictionary(search);
    }, 250);
    return () => clearTimeout(timer);
  }, [search]);

  const handleToggleFav = async (id: number) => {
    try {
      await api.toggleFavoriteWord(id);
      loadDictionary();
      showToast('Статус избранного обновлён ⭐', 'success');
    } catch (err) {
      console.error(err);
    }
  };

  const handleDelete = async (id: number, wordTitle: string) => {
    try {
      await api.deleteDictionaryWord(id);
      loadDictionary();
      showToast(`Слово «${wordTitle}» удалено из словаря`, 'info');
    } catch (err) {
      showToast('Ошибка при удалении слова', 'error');
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
      const addedWord = newWord.trim();
      setNewWord('');
      setNewTranslation('');
      setNewContext('');
      loadDictionary();
      showToast(`Слово «${addedWord}» успешно добавлено в словарь! 🎉`, 'success');
    } catch (err: any) {
      showToast(err.message || 'Ошибка добавления слова', 'error');
    }
  };

  const handleNextTrainerCard = (isKnown: boolean) => {
    if (isKnown) {
      setKnownCount((prev) => prev + 1);
    }
    if (trainerIndex + 1 >= entries.length) {
      // Completed stack celebration
      confetti({
        particleCount: 70,
        spread: 60,
        origin: { y: 0.6 },
        colors: ['#0C2340', '#1E50A0', '#F59E0B', '#10B981'],
      });
      showToast(`Тренировка завершена! Вы повторили ${entries.length} карточек 🎉`, 'success');
      setIsTrainerOpen(false);
      setTrainerIndex(0);
      setIsFlipped(false);
      return;
    }
    setTrainerIndex((prev) => prev + 1);
    setIsFlipped(false);
  };

  return (
    <div className="space-y-6 max-w-5xl mx-auto pb-16 relative">
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
                setKnownCount(0);
                setIsTrainerOpen(true);
              }}
              className="px-4 py-2.5 rounded-full bg-[#1E50A0] hover:bg-[#2563EB] text-white font-bold text-xs shadow-md flex items-center gap-2 transition-all shrink-0 cursor-pointer"
            >
              <RotateCw className="w-4 h-4" />
              <span>Режим карточек</span>
            </button>
          )}

          <button
            onClick={() => setAddModalOpen(true)}
            className="px-4 py-2.5 rounded-full bg-white text-[#0C2340] hover:bg-slate-100 font-bold text-xs shadow-md flex items-center gap-1.5 transition-all shrink-0 cursor-pointer"
          >
            <Plus className="w-4 h-4" />
            <span>Добавить слово</span>
          </button>
        </div>
      </div>

      {/* Search & Favorites Filter Bar */}
      <div className="bg-white rounded-2xl p-4 border border-slate-200/90 shadow-sm space-y-3">
        <div className="flex flex-col sm:flex-row items-center justify-between gap-3">
          <div className="relative flex-1 w-full">
            <Search className="w-4 h-4 text-slate-400 absolute left-3.5 top-3" />
            <input
              type="text"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder="Поиск по словам и переводам..."
              className="w-full pl-10 pr-10 py-2 rounded-full bg-[#F1F5F9] border-none text-xs sm:text-sm text-slate-900 focus:outline-none focus:ring-2 focus:ring-[#1E50A0]"
            />
            {search && (
              <button
                type="button"
                onClick={() => setSearch('')}
                className="absolute right-3.5 top-2.5 text-slate-400 hover:text-slate-600"
              >
                <X className="w-4 h-4" />
              </button>
            )}
          </div>

          <button
            onClick={() => setFavoriteOnly(!favoriteOnly)}
            className={`px-4 py-2 rounded-full text-xs font-bold flex items-center gap-2 transition-colors shrink-0 w-full sm:w-auto justify-center cursor-pointer ${
              favoriteOnly
                ? 'bg-amber-100 text-amber-800 border border-amber-300'
                : 'bg-slate-100 text-slate-600 hover:bg-slate-200'
            }`}
          >
            <Star className={`w-3.5 h-3.5 ${favoriteOnly ? 'fill-amber-600' : ''}`} />
            <span>Только избранные</span>
          </button>
        </div>

        {/* Bashkir Quick Letters Pill Bar */}
        <div className="flex items-center gap-1.5 overflow-x-auto pb-1 pt-1 border-t border-slate-100">
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
      </div>

      {/* Dictionary Items List */}
      {loading ? (
        <div className="text-center py-12 text-slate-400 text-sm">
          Загрузка словаря...
        </div>
      ) : entries.length === 0 ? (
        <div className="bg-white rounded-3xl p-12 text-center border border-slate-200 text-slate-400 text-sm space-y-3">
          <BookA className="w-10 h-10 mx-auto text-slate-300" />
          <p className="font-semibold text-slate-600">
            {search ? 'По вашему запросу слов не найдено' : 'В словаре пока нет слов'}
          </p>
          <p className="text-xs text-slate-400">
            {search
              ? 'Попробуйте стереть запрос или добавить новое слово в словарь'
              : 'Добавьте свои первые слова или сохраняйте их из ежедневных заданий!'}
          </p>
          {search && (
            <button
              onClick={() => setSearch('')}
              className="px-4 py-2 rounded-full bg-[#1E50A0] text-white text-xs font-bold hover:bg-blue-700 transition-colors cursor-pointer"
            >
              Сбросить поиск
            </button>
          )}
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {entries.map((item) => (
            <div
              key={item.id}
              className="bg-white rounded-3xl p-5 border border-slate-200/90 shadow-sm hover:shadow-md transition-shadow flex flex-col justify-between space-y-3"
            >
              <div className="space-y-2">
                <div className="flex items-start justify-between gap-2">
                  <div className="flex items-center gap-2">
                    <span className="text-lg font-black text-[#0C2340] font-serif">
                      {item.word}
                    </span>
                    <AudioSpeakerButton text={item.word} size="sm" />
                  </div>

                  <div className="flex items-center gap-1 shrink-0">
                    <button
                      onClick={() => handleToggleFav(item.id)}
                      className="p-1.5 rounded-lg text-slate-300 hover:text-amber-500 transition-colors cursor-pointer"
                      title={item.is_favorite ? 'Убрать из избранного' : 'В избранное'}
                    >
                      <Star className={`w-4 h-4 ${item.is_favorite ? 'fill-amber-500 text-amber-500' : ''}`} />
                    </button>
                    <button
                      onClick={() => handleDelete(item.id, item.word)}
                      className="p-1.5 rounded-lg text-slate-300 hover:text-red-500 transition-colors cursor-pointer"
                      title="Удалить из словаря"
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

      {/* Leitner Spaced Repetition Flashcard Trainer Modal */}
      {isTrainerOpen && entries.length > 0 && (
        <div className="fixed inset-0 z-50 bg-black/60 backdrop-blur-xs flex items-center justify-center p-4">
          <div className="bg-white rounded-3xl max-w-md w-full p-6 text-center shadow-2xl border border-slate-200 space-y-6 animate-scale-up">
            <div className="flex items-center justify-between text-xs text-slate-400">
              <span className="font-bold text-[#0C2340]">Карточка {trainerIndex + 1} из {entries.length}</span>
              <button
                onClick={() => setIsTrainerOpen(false)}
                className="font-bold text-slate-400 hover:text-slate-600 text-lg cursor-pointer"
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

            {/* Audio pronunciation for flashcard */}
            <div className="flex items-center justify-center pt-1">
              <AudioSpeakerButton
                text={entries[trainerIndex].word}
                size="md"
                showLabel={true}
                label="Прослушать произношение"
              />
            </div>

            {/* Leitner Spaced Repetition action buttons */}
            <div className="grid grid-cols-2 gap-3 pt-2">
              <button
                onClick={() => handleNextTrainerCard(false)}
                className="py-3 px-4 rounded-2xl bg-amber-50 hover:bg-amber-100 text-amber-800 font-bold text-xs border border-amber-200/80 transition-colors cursor-pointer flex items-center justify-center gap-1.5"
              >
                <span>🔴 Повторить</span>
              </button>
              <button
                onClick={() => handleNextTrainerCard(true)}
                className="py-3 px-4 rounded-2xl bg-emerald-50 hover:bg-emerald-100 text-emerald-800 font-bold text-xs border border-emerald-200/80 transition-colors cursor-pointer flex items-center justify-center gap-1.5"
              >
                <span>🟢 Знаю</span>
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Add Word Modal with Virtual Keyboard */}
      {addModalOpen && (
        <div className="fixed inset-0 z-50 bg-black/60 backdrop-blur-xs flex items-center justify-center p-4">
          <form
            onSubmit={handleAddSubmit}
            className="bg-white rounded-3xl max-w-md w-full p-6 space-y-4 shadow-2xl border border-slate-200 animate-scale-up"
          >
            <div className="flex items-center justify-between pb-2 border-b border-slate-100">
              <h3 className="font-extrabold text-base text-[#0C2340]">Добавить слово в словарь</h3>
              <button
                type="button"
                onClick={() => setAddModalOpen(false)}
                className="text-slate-400 hover:text-slate-600 font-bold text-lg cursor-pointer"
              >
                ✕
              </button>
            </div>

            <div className="space-y-3">
              <div>
                <label className="block text-xs font-bold text-slate-700 mb-1">
                  Слово на башкирском языке *
                </label>
                <input
                  type="text"
                  required
                  value={newWord}
                  onChange={(e) => setNewWord(e.target.value)}
                  placeholder="например, Ҡояш"
                  className="w-full px-4 py-2.5 rounded-xl border border-slate-300 text-sm focus:outline-none focus:ring-2 focus:ring-[#1E50A0]"
                />
                {/* Virtual Letters in Add Modal */}
                <div className="flex items-center gap-1 overflow-x-auto pt-2">
                  {BASHKIR_LETTERS.map((letter) => (
                    <button
                      key={letter}
                      type="button"
                      onClick={() => setNewWord((prev) => prev + letter.toLowerCase())}
                      className="w-6 h-6 rounded bg-slate-100 hover:bg-blue-100 text-[11px] font-bold text-slate-700 shrink-0 cursor-pointer"
                      title={`Вставить ${letter}`}
                    >
                      {letter}
                    </button>
                  ))}
                </div>
              </div>

              <div>
                <label className="block text-xs font-bold text-slate-700 mb-1">
                  Перевод на русский *
                </label>
                <input
                  type="text"
                  required
                  value={newTranslation}
                  onChange={(e) => setNewTranslation(e.target.value)}
                  placeholder="например, Солнце"
                  className="w-full px-4 py-2.5 rounded-xl border border-slate-300 text-sm focus:outline-none focus:ring-2 focus:ring-[#1E50A0]"
                />
              </div>

              <div>
                <label className="block text-xs font-bold text-slate-700 mb-1">
                  Контекст / пример употребления
                </label>
                <textarea
                  value={newContext}
                  onChange={(e) => setNewContext(e.target.value)}
                  placeholder="например, Яҡты ҡояш балҡый"
                  rows={2}
                  className="w-full px-4 py-2 rounded-xl border border-slate-300 text-xs focus:outline-none focus:ring-2 focus:ring-[#1E50A0]"
                />
              </div>
            </div>

            <div className="pt-2 flex items-center justify-end gap-2">
              <button
                type="button"
                onClick={() => setAddModalOpen(false)}
                className="px-4 py-2 rounded-full border border-slate-200 text-xs font-semibold text-slate-600 hover:bg-slate-50 cursor-pointer"
              >
                Отмена
              </button>
              <button
                type="submit"
                className="px-5 py-2 rounded-full bg-[#1E50A0] hover:bg-blue-700 text-white text-xs font-bold shadow-md cursor-pointer"
              >
                Сохранить в словарь
              </button>
            </div>
          </form>
        </div>
      )}
    </div>
  );
};
