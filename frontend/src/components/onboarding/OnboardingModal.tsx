import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../../context/AuthContext';
import { api } from '../../api/client';
import { SchoolOption } from '../../types';
import { BashkirOrnament } from '../ui/BashkirOrnament';
import { playChime } from '../../utils/speech';
import confetti from 'canvas-confetti';
import {
  Sparkles,
  ArrowRight,
  Check,
  School,
  Languages,
  BookOpen,
  Award,
  Coins,
  Shield,
  CheckCircle2
} from 'lucide-react';

interface OnboardingModalProps {
  isOpen: boolean;
  onClose: () => void;
}

export const OnboardingModal: React.FC<OnboardingModalProps> = ({ isOpen, onClose }) => {
  const { user, refreshUser } = useAuth();
  const navigate = useNavigate();

  const [step, setStep] = useState(1);
  const [schools, setSchools] = useState<SchoolOption[]>([]);
  const [selectedLevel, setSelectedLevel] = useState<'beginner' | 'intermediate' | 'advanced'>('intermediate');
  const [selectedSchoolId, setSelectedSchoolId] = useState<number | undefined>(user?.school_id);
  const [selectedClassId, setSelectedClassId] = useState<number | undefined>(user?.class_id);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    if (isOpen) {
      api.getSchools().then(setSchools).catch(console.error);
    }
  }, [isOpen]);

  if (!isOpen) return null;

  const currentSchool = schools.find((s) => s.id === selectedSchoolId);

  const handleFinish = async () => {
    setSaving(true);
    try {
      if (selectedSchoolId !== user?.school_id || selectedClassId !== user?.class_id) {
        await api.updateProfile({
          school_id: selectedSchoolId,
          class_id: selectedClassId,
        });
        await refreshUser();
      }

      if (user) {
        localStorage.setItem(`onboarding_completed_${user.id}`, 'true');
      }

      playChime();
      confetti({
        particleCount: 70,
        spread: 80,
        origin: { y: 0.6 },
        colors: ['#0C2340', '#1E50A0', '#F59E0B', '#10B981'],
      });

      onClose();
      navigate('/task');
    } catch (err) {
      console.error(err);
      onClose();
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 bg-black/60 backdrop-blur-xs flex items-center justify-center p-4 select-none animate-fade-in">
      <div className="bg-white rounded-3xl max-w-lg w-full p-6 sm:p-8 shadow-2xl border border-slate-200 relative overflow-hidden space-y-6">
        {/* Top Progress indicator */}
        <div className="flex items-center justify-between border-b border-slate-100 pb-4">
          <div className="flex items-center gap-2">
            <span className="text-xs font-bold text-[#1E50A0]">Шаг {step} из 3</span>
            <div className="flex items-center gap-1.5">
              {[1, 2, 3].map((s) => (
                <div
                  key={s}
                  className={`w-6 h-1.5 rounded-full transition-all ${
                    s <= step ? 'bg-[#1E50A0]' : 'bg-slate-200'
                  }`}
                />
              ))}
            </div>
          </div>

          <button
            onClick={onClose}
            className="text-xs font-bold text-slate-400 hover:text-slate-600 transition-colors"
          >
            Пропустить
          </button>
        </div>

        {/* STEP 1: WELCOME & MISSION */}
        {step === 1 && (
          <div className="text-center space-y-4 animate-scale-up">
            <div className="w-16 h-16 rounded-3xl bg-[#0C2340] text-white flex items-center justify-center mx-auto shadow-md">
              <BashkirOrnament className="w-9 h-9 text-white" />
            </div>

            <div className="space-y-1.5">
              <h2 className="text-2xl font-black text-[#0C2340]">
                Хәйерле сәғәттә, Хранитель!
              </h2>
              <p className="text-xs sm:text-sm text-slate-500 leading-relaxed max-w-sm mx-auto">
                Добро пожаловать в <strong>ЙӘНЛӘНДЕР</strong> — народную платформу сохранения и верификации башкирского языка.
              </p>
            </div>

            <div className="p-4 rounded-2xl bg-blue-50/60 border border-blue-100 text-left space-y-2 text-xs">
              <div className="flex items-center gap-2 font-bold text-[#1E50A0]">
                <Sparkles className="w-4 h-4 text-amber-500 shrink-0" />
                <span>Ваша великая миссия:</span>
              </div>
              <ul className="space-y-1.5 text-slate-600 list-disc list-inside">
                <li>Проверяйте переводы литературных отрывков и пословиц.</li>
                <li>Зарабатывайте опыт (XP) и собирайте народную сокровищницу.</li>
                <li>Помогайте своей школе подниматься в общем рейтинге Башкортостана.</li>
              </ul>
            </div>

            <button
              onClick={() => setStep(2)}
              className="w-full py-3.5 rounded-full bg-[#0C2340] hover:bg-[#1A4480] text-white font-bold text-xs sm:text-sm shadow-md flex items-center justify-center gap-2 transition-all cursor-pointer"
            >
              <span>Продолжить</span>
              <ArrowRight className="w-4 h-4" />
            </button>
          </div>
        )}

        {/* STEP 2: LANGUAGE GOAL & LEVEL */}
        {step === 2 && (
          <div className="space-y-5 animate-scale-up">
            <div className="text-center space-y-1">
              <h3 className="text-xl font-black text-[#0C2340]">
                Ваш уровень владения языком
              </h3>
              <p className="text-xs text-slate-500">
                Языковая пара: <strong>Башҡортса ↔ Русский</strong>
              </p>
            </div>

            <div className="space-y-2.5">
              {[
                {
                  id: 'beginner',
                  title: 'Начинающий (Башлаусы)',
                  desc: 'Учу первые слова, хочу запомнить базовые пословицы и фразы.',
                  icon: BookOpen,
                },
                {
                  id: 'intermediate',
                  title: 'Знаток (Белеүсе)',
                  desc: 'Понимаю язык, могу проверять варианты и находить неточности.',
                  icon: Award,
                },
                {
                  id: 'advanced',
                  title: 'Мастер и Наставник (Оҫта)',
                  desc: 'Свободно владею башкирским, готов к сложным эпическим текстам.',
                  icon: Shield,
                },
              ].map((lvl) => {
                const Icon = lvl.icon;
                const isSelected = selectedLevel === lvl.id;
                return (
                  <div
                    key={lvl.id}
                    onClick={() => setSelectedLevel(lvl.id as any)}
                    className={`p-4 rounded-2xl border transition-all cursor-pointer flex items-center gap-3.5 ${
                      isSelected
                        ? 'border-[#1E50A0] bg-[#EEF2FF] shadow-xs'
                        : 'border-slate-200 hover:border-slate-300 bg-white'
                    }`}
                  >
                    <div
                      className={`w-10 h-10 rounded-xl flex items-center justify-center shrink-0 ${
                        isSelected ? 'bg-[#1E50A0] text-white' : 'bg-slate-100 text-slate-500'
                      }`}
                    >
                      <Icon className="w-5 h-5" />
                    </div>
                    <div className="flex-1">
                      <div className="font-extrabold text-xs sm:text-sm text-[#0C2340]">
                        {lvl.title}
                      </div>
                      <div className="text-[11px] text-slate-500 mt-0.5">{lvl.desc}</div>
                    </div>
                    {isSelected && <Check className="w-4 h-4 text-[#1E50A0] shrink-0" />}
                  </div>
                );
              })}
            </div>

            <div className="flex items-center gap-3 pt-2">
              <button
                onClick={() => setStep(1)}
                className="py-3 px-5 rounded-full border border-slate-200 text-xs font-bold text-slate-600 hover:bg-slate-50"
              >
                Назад
              </button>
              <button
                onClick={() => setStep(3)}
                className="flex-1 py-3 rounded-full bg-[#0C2340] hover:bg-[#1A4480] text-white font-bold text-xs sm:text-sm shadow-md flex items-center justify-center gap-2 transition-all cursor-pointer"
              >
                <span>Далее</span>
                <ArrowRight className="w-4 h-4" />
              </button>
            </div>
          </div>
        )}

        {/* STEP 3: SCHOOL SELECTION & CELEBRATION */}
        {step === 3 && (
          <div className="space-y-5 animate-scale-up">
            <div className="text-center space-y-1">
              <h3 className="text-xl font-black text-[#0C2340]">
                Школьная команда и посвящение
              </h3>
              <p className="text-xs text-slate-500">
                Каждое ваше выполненное задание приносит очки вашей школе.
              </p>
            </div>

            <div className="space-y-3">
              <div>
                <label className="block text-xs font-bold text-slate-700 mb-1">
                  Выберите учебное заведение
                </label>
                <select
                  value={selectedSchoolId || ''}
                  onChange={(e) => {
                    const id = e.target.value ? Number(e.target.value) : undefined;
                    setSelectedSchoolId(id);
                    setSelectedClassId(undefined);
                  }}
                  className="w-full px-4 py-2.5 rounded-2xl bg-slate-50 border border-slate-200 text-xs sm:text-sm font-semibold text-slate-800 focus:outline-none focus:ring-2 focus:ring-[#1E50A0]"
                >
                  <option value="">Индивидуальный Хранитель (без школы)</option>
                  {schools.map((s) => (
                    <option key={s.id} value={s.id}>
                      {s.name} ({s.city})
                    </option>
                  ))}
                </select>
              </div>

              {currentSchool && currentSchool.classes && currentSchool.classes.length > 0 && (
                <div>
                  <label className="block text-xs font-bold text-slate-700 mb-1">Класс</label>
                  <select
                    value={selectedClassId || ''}
                    onChange={(e) => setSelectedClassId(e.target.value ? Number(e.target.value) : undefined)}
                    className="w-full px-4 py-2.5 rounded-2xl bg-slate-50 border border-slate-200 text-xs sm:text-sm font-semibold text-slate-800 focus:outline-none focus:ring-2 focus:ring-[#1E50A0]"
                  >
                    <option value="">Выберите класс</option>
                    {currentSchool.classes.map((c) => (
                      <option key={c.id} value={c.id}>
                        Класс {c.name}
                      </option>
                    ))}
                  </select>
                </div>
              )}
            </div>

            {/* Gift Box Preview */}
            <div className="p-4 rounded-2xl bg-amber-50 border border-amber-200 text-center space-y-1">
              <div className="text-xs font-bold text-amber-900 flex items-center justify-center gap-1.5">
                <Coins className="w-4 h-4 text-amber-600 animate-pulse" />
                <span>Приветственный подарок новому Хранителю:</span>
              </div>
              <div className="text-sm font-black text-amber-700">
                +50 Опыта (XP) и +25 монет «Ҡурай»!
              </div>
            </div>

            <div className="flex items-center gap-3 pt-2">
              <button
                onClick={() => setStep(2)}
                className="py-3 px-5 rounded-full border border-slate-200 text-xs font-bold text-slate-600 hover:bg-slate-50"
              >
                Назад
              </button>
              <button
                onClick={handleFinish}
                disabled={saving}
                className="flex-1 py-3.5 rounded-full bg-[#0C2340] hover:bg-[#1A4480] text-white font-bold text-xs sm:text-sm shadow-md flex items-center justify-center gap-2 transition-all cursor-pointer disabled:opacity-50"
              >
                <Sparkles className="w-4 h-4 text-amber-400" />
                <span>{saving ? 'Посвящение...' : 'Начать путь Хранителя'}</span>
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};
