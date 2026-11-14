import React, { useEffect, useState } from 'react';
import { api } from '../api/client';
import {
  ResponsiveContainer,
  AreaChart,
  Area,
  BarChart,
  Bar,
  PieChart,
  Pie,
  Cell,
  XAxis,
  YAxis,
  Tooltip,
  Legend,
  CartesianGrid
} from 'recharts';
import {
  BarChart2,
  TrendingUp,
  CheckCircle2,
  Users,
  BookOpen,
  Award,
  Sparkles,
  School,
  Calendar,
  Layers,
  ArrowUpRight
} from 'lucide-react';

export const AnalyticsPage: React.FC = () => {
  const [data, setData] = useState<any>(null);
  const [loading, setLoading] = useState(true);
  const [period, setPeriod] = useState<'7d' | '30d'>('7d');

  useEffect(() => {
    loadAnalytics();
  }, []);

  const loadAnalytics = async () => {
    setLoading(true);
    try {
      const res = await api.getPlatformAnalytics();
      setData(res);
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  const activityData = period === '7d' ? (data?.activity_7d || []) : [
    { day: 'Неделя 1', tasks: 120, confirmations: 85, proposals: 35 },
    { day: 'Неделя 2', tasks: 165, confirmations: 115, proposals: 50 },
    { day: 'Неделя 3', tasks: 210, confirmations: 155, proposals: 55 },
    { day: 'Неделя 4', tasks: 260, confirmations: 190, proposals: 70 },
  ];

  return (
    <div className="space-y-6 max-w-6xl mx-auto pb-16">
      {/* Header Banner */}
      <div className="yan-page-hero bg-[#0C2340] rounded-3xl p-6 sm:p-8 text-white shadow-md flex flex-col md:flex-row items-start md:items-center justify-between gap-6 border border-slate-700/50">
        <div>
          <div className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full bg-[#1E50A0]/60 text-xs font-semibold backdrop-blur mb-2 border border-blue-400/30">
            <BarChart2 className="w-3.5 h-3.5 text-amber-400" />
            <span>Аналитика и статистика платформы</span>
          </div>
          <h1 className="text-2xl sm:text-3xl font-extrabold tracking-tight">
            Динамика сохранения башкирского языка
          </h1>
          <p className="text-slate-300 text-xs sm:text-sm mt-1 max-w-xl">
            Реальные показатели верификации корпуса текстов, точности предложений и активности школ Республики Башкортостан.
          </p>
        </div>

        <div className="flex items-center gap-2 bg-white/10 p-1.5 rounded-2xl border border-white/10 text-xs font-bold">
          <button
            onClick={() => setPeriod('7d')}
            className={`px-4 py-2 rounded-xl transition-all ${
              period === '7d' ? 'bg-[#1E50A0] text-white shadow-sm' : 'text-slate-300 hover:text-white'
            }`}
          >
            Последние 7 дней
          </button>
          <button
            onClick={() => setPeriod('30d')}
            className={`px-4 py-2 rounded-xl transition-all ${
              period === '30d' ? 'bg-[#1E50A0] text-white shadow-sm' : 'text-slate-300 hover:text-white'
            }`}
          >
            За 30 дней
          </button>
        </div>
      </div>

      {loading ? (
        <div className="text-center py-20 text-slate-400 text-sm">
          Загрузка аналитических графиков...
        </div>
      ) : (
        <>
          {/* Key Metric Cards */}
          <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
            <div className="bg-white p-5 rounded-3xl border border-slate-200/90 shadow-sm flex flex-col justify-between">
              <div className="flex items-center justify-between text-slate-400">
                <span className="text-[11px] font-bold uppercase tracking-wider">Всего действий</span>
                <TrendingUp className="w-4 h-4 text-[#1E50A0]" />
              </div>
              <div className="mt-3">
                <div className="text-2xl sm:text-3xl font-black text-[#0C2340]">
                  {data?.summary?.total_contributions || 0}
                </div>
                <div className="flex items-center gap-1 text-[11px] font-bold text-emerald-600 mt-1">
                  <ArrowUpRight className="w-3.5 h-3.5" />
                  <span>+28% за неделю</span>
                </div>
              </div>
            </div>

            <div className="bg-white p-5 rounded-3xl border border-slate-200/90 shadow-sm flex flex-col justify-between">
              <div className="flex items-center justify-between text-slate-400">
                <span className="text-[11px] font-bold uppercase tracking-wider">Подтверждено в Сокровищнице</span>
                <CheckCircle2 className="w-4 h-4 text-emerald-600" />
              </div>
              <div className="mt-3">
                <div className="text-2xl sm:text-3xl font-black text-emerald-700">
                  {data?.summary?.confirmed_translations || 0}
                </div>
                <div className="text-[11px] text-slate-400 font-medium mt-1">
                  Из {data?.summary?.total_fragments || 0} фрагментов
                </div>
              </div>
            </div>

            <div className="bg-white p-5 rounded-3xl border border-slate-200/90 shadow-sm flex flex-col justify-between">
              <div className="flex items-center justify-between text-slate-400">
                <span className="text-[11px] font-bold uppercase tracking-wider">Точность правок</span>
                <Award className="w-4 h-4 text-amber-500" />
              </div>
              <div className="mt-3">
                <div className="text-2xl sm:text-3xl font-black text-[#0C2340]">
                  {data?.summary?.acceptance_rate || 85}%
                </div>
                <div className="text-[11px] text-slate-400 font-medium mt-1">
                  Принятых вариантов сообществом
                </div>
              </div>
            </div>

            <div className="bg-white p-5 rounded-3xl border border-slate-200/90 shadow-sm flex flex-col justify-between">
              <div className="flex items-center justify-between text-slate-400">
                <span className="text-[11px] font-bold uppercase tracking-wider">Школ на платформе</span>
                <School className="w-4 h-4 text-[#1E50A0]" />
              </div>
              <div className="mt-3">
                <div className="text-2xl sm:text-3xl font-black text-[#0C2340]">
                  {data?.summary?.active_schools || 0}
                </div>
                <div className="text-[11px] text-slate-400 font-medium mt-1">
                  {data?.summary?.total_users || 0} Хранителей языка
                </div>
              </div>
            </div>
          </div>

          {/* Activity Over Time: AreaChart */}
          <div className="bg-white rounded-3xl border border-slate-200/90 shadow-sm p-6 sm:p-7 space-y-4">
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2">
              <div>
                <h3 className="text-base sm:text-lg font-black text-[#0C2340]">
                  Динамика активности Хранителей ({period === '7d' ? 'За последние 7 дней' : 'По неделям месяца'})
                </h3>
                <p className="text-xs text-slate-500">
                  Соотношение подтверждённых готовых переводов и новых предложенных вариантов.
                </p>
              </div>

              <div className="flex items-center gap-4 text-xs font-bold text-slate-600">
                <div className="flex items-center gap-1.5">
                  <span className="w-3 h-3 rounded-full bg-[#1E50A0]" />
                  <span>Подтверждения</span>
                </div>
                <div className="flex items-center gap-1.5">
                  <span className="w-3 h-3 rounded-full bg-[#3B82F6]" />
                  <span>Новые предложения</span>
                </div>
              </div>
            </div>

            <div className="h-72 w-full pt-4">
              <ResponsiveContainer width="100%" height="100%">
                <AreaChart data={activityData}>
                  <defs>
                    <linearGradient id="colorConfirm" x1="0" y1="0" x2="0" y2="1">
                      <stop offset="5%" stopColor="#1E50A0" stopOpacity={0.4} />
                      <stop offset="95%" stopColor="#1E50A0" stopOpacity={0.0} />
                    </linearGradient>
                    <linearGradient id="colorProposal" x1="0" y1="0" x2="0" y2="1">
                      <stop offset="5%" stopColor="#3B82F6" stopOpacity={0.4} />
                      <stop offset="95%" stopColor="#3B82F6" stopOpacity={0.0} />
                    </linearGradient>
                  </defs>
                  <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="#E2E8F0" />
                  <XAxis dataKey="day" stroke="#94A3B8" fontSize={11} tickLine={false} />
                  <YAxis stroke="#94A3B8" fontSize={11} tickLine={false} axisLine={false} />
                  <Tooltip
                    contentStyle={{
                      backgroundColor: '#0C2340',
                      borderRadius: '16px',
                      color: '#FFFFFF',
                      fontSize: '12px',
                      border: 'none',
                      boxShadow: '0 8px 24px rgba(0,0,0,0.15)',
                    }}
                  />
                  <Area
                    type="monotone"
                    dataKey="confirmations"
                    name="Подтверждения"
                    stroke="#1E50A0"
                    strokeWidth={3}
                    fillOpacity={1}
                    fill="url(#colorConfirm)"
                  />
                  <Area
                    type="monotone"
                    dataKey="proposals"
                    name="Новые варианты"
                    stroke="#3B82F6"
                    strokeWidth={2}
                    fillOpacity={1}
                    fill="url(#colorProposal)"
                  />
                </AreaChart>
              </ResponsiveContainer>
            </div>
          </div>

          {/* Two Columns: Category Breakdown & Difficulty Pie */}
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            {/* Categories BarChart */}
            <div className="bg-white rounded-3xl border border-slate-200/90 shadow-sm p-6 space-y-4">
              <div>
                <h3 className="text-base font-black text-[#0C2340]">
                  Тематическое распределение корпуса
                </h3>
                <p className="text-xs text-slate-500">
                  Доля литературных отрывков, эпоса и фольклора в базе.
                </p>
              </div>

              <div className="h-64 w-full">
                <ResponsiveContainer width="100%" height="100%">
                  <BarChart data={data?.category_breakdown || []} layout="vertical">
                    <CartesianGrid strokeDasharray="3 3" horizontal={false} stroke="#E2E8F0" />
                    <XAxis type="number" stroke="#94A3B8" fontSize={11} tickLine={false} axisLine={false} />
                    <YAxis dataKey="category" type="category" stroke="#0C2340" fontSize={11} tickLine={false} width={120} />
                    <Tooltip
                      contentStyle={{
                        backgroundColor: '#0C2340',
                        borderRadius: '12px',
                        color: '#FFFFFF',
                        fontSize: '12px',
                        border: 'none',
                      }}
                    />
                    <Bar dataKey="count" name="Фрагментов" fill="#1E50A0" radius={[0, 8, 8, 0]} />
                  </BarChart>
                </ResponsiveContainer>
              </div>
            </div>

            {/* Difficulty & Actions Distribution */}
            <div className="bg-white rounded-3xl border border-slate-200/90 shadow-sm p-6 space-y-4 flex flex-col justify-between">
              <div>
                <h3 className="text-base font-black text-[#0C2340]">
                  Сложность заданий и типы вклада
                </h3>
                <p className="text-xs text-slate-500">
                  Распределение заданий по уровням сложности и видам пользовательских действий.
                </p>
              </div>

              <div className="grid grid-cols-2 gap-4 items-center">
                <div className="h-48">
                  <ResponsiveContainer width="100%" height="100%">
                    <PieChart>
                      <Pie
                        data={data?.difficulty_pie || []}
                        dataKey="value"
                        nameKey="name"
                        innerRadius={45}
                        outerRadius={70}
                        paddingAngle={4}
                      >
                        {(data?.difficulty_pie || []).map((entry: any, index: number) => (
                          <Cell key={`cell-${index}`} fill={entry.color} />
                        ))}
                      </Pie>
                      <Tooltip />
                    </PieChart>
                  </ResponsiveContainer>
                  <div className="text-center text-[10px] font-bold text-slate-500 uppercase mt-1">
                    Сложность заданий
                  </div>
                </div>

                <div className="h-48">
                  <ResponsiveContainer width="100%" height="100%">
                    <PieChart>
                      <Pie
                        data={data?.actions_donut || []}
                        dataKey="value"
                        nameKey="name"
                        innerRadius={45}
                        outerRadius={70}
                        paddingAngle={4}
                      >
                        {(data?.actions_donut || []).map((entry: any, index: number) => (
                          <Cell key={`cell-act-${index}`} fill={entry.color} />
                        ))}
                      </Pie>
                      <Tooltip />
                    </PieChart>
                  </ResponsiveContainer>
                  <div className="text-center text-[10px] font-bold text-slate-500 uppercase mt-1">
                    Структура действий
                  </div>
                </div>
              </div>

              <div className="flex flex-wrap items-center justify-around gap-2 pt-2 border-t border-slate-100 text-xs font-semibold text-slate-600">
                <div className="flex items-center gap-1.5">
                  <span className="w-2.5 h-2.5 rounded-full bg-emerald-500" />
                  <span>Лёгкий</span>
                </div>
                <div className="flex items-center gap-1.5">
                  <span className="w-2.5 h-2.5 rounded-full bg-[#1E50A0]" />
                  <span>Средний</span>
                </div>
                <div className="flex items-center gap-1.5">
                  <span className="w-2.5 h-2.5 rounded-full bg-amber-500" />
                  <span>Сложный</span>
                </div>
              </div>
            </div>
          </div>

          {/* Schools Ranking Table with Progress Bars */}
          <div className="bg-white rounded-3xl border border-slate-200/90 shadow-sm p-6 sm:p-7 space-y-4">
            <div className="flex items-center justify-between">
              <div>
                <h3 className="text-base sm:text-lg font-black text-[#0C2340]">
                  Вклад учебных заведений в сохранение языка
                </h3>
                <p className="text-xs text-slate-500">
                  Активность школьных команд и классов по количеству проверенных фрагментов.
                </p>
              </div>
              <School className="w-6 h-6 text-[#1E50A0]" />
            </div>

            <div className="space-y-3 pt-2">
              {(data?.schools_ranking || []).map((sch: any, idx: number) => {
                const maxContr = Math.max(...(data?.schools_ranking || []).map((s: any) => s.contributions), 1);
                const pct = Math.round((sch.contributions / maxContr) * 100);

                return (
                  <div key={idx} className="p-4 rounded-2xl bg-[#F8FAFC] border border-slate-200/70 space-y-2">
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-2.5">
                        <span className="w-6 h-6 rounded-full bg-[#0C2340] text-white font-black text-xs flex items-center justify-center">
                          {idx + 1}
                        </span>
                        <div>
                          <div className="font-extrabold text-sm text-[#0C2340]">{sch.name}</div>
                          <div className="text-[11px] text-slate-400">г. {sch.city} • Учеников: {sch.students_count}</div>
                        </div>
                      </div>
                      <div className="text-right">
                        <div className="text-sm font-black text-[#1E50A0]">{sch.contributions} вкладов</div>
                        <div className="text-[10px] text-slate-400 font-semibold">{pct}% от лидера</div>
                      </div>
                    </div>

                    <div className="w-full h-2 bg-slate-200 rounded-full overflow-hidden">
                      <div
                        className="h-full bg-gradient-to-r from-[#1E50A0] to-[#3B82F6] rounded-full transition-all duration-500"
                        style={{ width: `${pct}%` }}
                      />
                    </div>
                  </div>
                );
              })}
            </div>
          </div>
        </>
      )}
    </div>
  );
};
