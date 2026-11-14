import React, { useEffect, useState } from 'react';
import { api } from '../api/client';
import { useAuth } from '../context/AuthContext';
import { ShopItem } from '../types';
import { UserAvatar } from '../components/ui/UserAvatar';
import { BashkirOrnament } from '../components/ui/BashkirOrnament';
import { playChime } from '../utils/speech';
import confetti from 'canvas-confetti';
import {
  ShoppingBag,
  Coins,
  Shield,
  Sparkles,
  Sun,
  Gem,
  Award,
  Scroll,
  Check,
  CheckCircle2,
  AlertCircle
} from 'lucide-react';

export const ShopPage: React.FC = () => {
  const { user, refreshUser } = useAuth();
  const [items, setItems] = useState<ShopItem[]>([]);
  const [currency, setCurrency] = useState<number>(user?.stats?.currency || 0);
  const [category, setCategory] = useState<'all' | 'frame' | 'utility' | 'lore' | 'title'>('all');
  const [loading, setLoading] = useState(true);
  const [actionInProgress, setActionInProgress] = useState<string | null>(null);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  const loadShop = async () => {
    setLoading(true);
    try {
      const data = await api.getShopItems();
      setItems(data.items);
      setCurrency(data.currency);
    } catch (err: any) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadShop();
  }, []);

  const handleBuy = async (item: ShopItem) => {
    setActionInProgress(item.id);
    setSuccessMessage(null);
    setErrorMessage(null);

    try {
      const res = await api.buyShopItem(item.id);
      if (res.success) {
        setSuccessMessage(res.message);
        setCurrency(res.currency);
        playChime();
        confetti({
          particleCount: 50,
          spread: 60,
          origin: { y: 0.7 },
          colors: ['#0C2340', '#1E50A0', '#F59E0B', '#10B981'],
        });
        await refreshUser();
        await loadShop();
      } else {
        setErrorMessage(res.message);
      }
    } catch (err: any) {
      setErrorMessage(err.message || 'Ошибка при покупке');
    } finally {
      setActionInProgress(null);
    }
  };

  const handleEquip = async (item: ShopItem) => {
    setActionInProgress(item.id);
    setSuccessMessage(null);
    setErrorMessage(null);

    try {
      const categoryParam = item.category === 'title' ? 'title' : 'frame';
      await api.equipShopItem(item.id, categoryParam);
      await refreshUser();
      await loadShop();
    } catch (err: any) {
      setErrorMessage(err.message || 'Ошибка экипировки');
    } finally {
      setActionInProgress(null);
    }
  };

  const getIcon = (iconName: string) => {
    switch (iconName) {
      case 'shield':
        return <Shield className="w-6 h-6 text-emerald-600" />;
      case 'sun':
        return <Sun className="w-6 h-6 text-amber-500" />;
      case 'sparkles':
        return <Sparkles className="w-6 h-6 text-cyan-500" />;
      case 'gem':
        return <Gem className="w-6 h-6 text-emerald-500" />;
      case 'award':
        return <Award className="w-6 h-6 text-yellow-600" />;
      case 'scroll':
        return <Scroll className="w-6 h-6 text-blue-500" />;
      default:
        return <ShoppingBag className="w-6 h-6 text-[#1E50A0]" />;
    }
  };

  const filteredItems = items.filter((item) => {
    if (category === 'all') return true;
    return item.category === category;
  });

  return (
    <div className="space-y-8 max-w-5xl mx-auto pb-16">
      {/* Header Banner */}
      <div className="yan-page-hero bg-[#0C2340] rounded-3xl p-6 sm:p-8 text-white shadow-md flex flex-col md:flex-row items-start md:items-center justify-between gap-6 border border-slate-700/50">
        <div className="space-y-2">
          <div className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full bg-[#1E50A0]/60 text-xs font-semibold backdrop-blur border border-blue-400/30">
            <ShoppingBag className="w-3.5 h-3.5 text-amber-400" />
            <span>Лавка Хранителя языка</span>
          </div>
          <h1 className="text-2xl sm:text-3xl font-extrabold tracking-tight">
            Артефакты, Обереги и Оформление
          </h1>
          <p className="text-slate-300 text-xs sm:text-sm max-w-xl">
            Обменивайте заработанные монеты Ҡурай на обереги стрика, редкую народную мудрость и золотые оправы аватара.
          </p>
        </div>

        {/* Currency & Live Avatar Preview */}
        <div className="yan-balance-panel flex items-center gap-4 bg-white/10 backdrop-blur-md px-5 py-3.5 rounded-2xl border border-white/15">
          <UserAvatar user={user} size="lg" />
          <div>
            <div className="text-[11px] font-bold text-slate-300 uppercase tracking-wider">
              Ваш баланс
            </div>
            <div className="flex items-center gap-1.5 text-xl font-black text-amber-300">
              <Coins className="w-5 h-5 text-amber-500" />
              <span>{currency} Ҡурай</span>
            </div>
            <div className="text-[10px] text-slate-300">
              {user?.stats?.streak_protected ? 'Защита серии активна' : 'Серия без защиты'}
            </div>
          </div>
        </div>
      </div>

      {/* Notifications */}
      {successMessage && (
        <div className="bg-emerald-50 border border-emerald-200 text-emerald-800 text-xs p-4 rounded-2xl flex items-center gap-2 shadow-xs">
          <CheckCircle2 className="w-4 h-4 text-emerald-600 shrink-0" />
          <span className="font-semibold">{successMessage}</span>
        </div>
      )}

      {errorMessage && (
        <div className="bg-red-50 border border-red-200 text-red-700 text-xs p-4 rounded-2xl flex items-center gap-2 shadow-xs">
          <AlertCircle className="w-4 h-4 text-red-600 shrink-0" />
          <span className="font-semibold">{errorMessage}</span>
        </div>
      )}

      {/* Category Pills */}
      <div className="flex items-center gap-2 overflow-x-auto pb-1">
        {[
          { id: 'all', label: 'Все артефакты' },
          { id: 'frame', label: 'Рамки аватара' },
          { id: 'utility', label: 'Обереги серии' },
          { id: 'title', label: 'Титулы и Знаки' },
          { id: 'lore', label: 'Мудрость эпоса' },
        ].map((tab) => (
          <button
            key={tab.id}
            onClick={() => setCategory(tab.id as any)}
            className={`px-4 py-2 rounded-full text-xs font-bold whitespace-nowrap transition-all ${
              category === tab.id
                ? 'bg-[#1E50A0] text-white shadow-sm'
                : 'bg-white border border-slate-200 text-slate-700 hover:bg-slate-50'
            }`}
          >
            {tab.label}
          </button>
        ))}
      </div>

      {/* Items Grid */}
      {loading ? (
        <div className="text-center py-16 text-slate-400 text-sm">
          Загрузка ассортимента Лавки...
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5">
          {filteredItems.map((item) => {
            const canAfford = currency >= item.price;
            const isActing = actionInProgress === item.id;

            return (
              <div
                key={item.id}
                className="bg-white rounded-3xl p-6 border border-slate-200/90 shadow-sm hover:shadow-md transition-all flex flex-col justify-between space-y-4"
              >
                <div className="space-y-3">
                  <div className="flex items-center justify-between">
                    <div className="w-12 h-12 rounded-2xl bg-slate-50 border border-slate-100 flex items-center justify-center">
                      {getIcon(item.icon)}
                    </div>
                    <div className="flex items-center gap-1 px-3 py-1 rounded-full bg-amber-50 text-amber-700 font-bold text-xs border border-amber-200/60">
                      <Coins className="w-3.5 h-3.5 text-amber-500" />
                      <span>{item.price} Ҡурай</span>
                    </div>
                  </div>

                  <div>
                    <h3 className="font-extrabold text-base text-[#0C2340]">
                      {item.name}
                    </h3>
                    <p className="text-xs text-slate-500 mt-1 leading-relaxed">
                      {item.description}
                    </p>
                  </div>
                </div>

                <div className="pt-3 border-t border-slate-100 flex items-center justify-between gap-3">
                  {item.owned ? (
                    item.category === 'utility' ? (
                      <div className="w-full py-2.5 px-4 rounded-full bg-emerald-50 border border-emerald-200 text-emerald-700 font-bold text-xs flex items-center justify-center gap-1.5">
                        <Check className="w-3.5 h-3.5 text-emerald-600" />
                        <span>Активно на аккаунте</span>
                      </div>
                    ) : item.equipped ? (
                      <button
                        onClick={() => handleEquip(item)}
                        disabled={isActing}
                        className="w-full py-2.5 px-4 rounded-full bg-emerald-50 border border-emerald-200 text-emerald-700 font-bold text-xs flex items-center justify-center gap-1.5 hover:bg-emerald-100 transition-colors cursor-pointer"
                        title="Нажмите, чтобы снять и вернуть обычный вид"
                      >
                        <Check className="w-3.5 h-3.5 text-emerald-600" />
                        <span>Надето (нажмите, чтобы снять)</span>
                      </button>
                    ) : (
                      <button
                        onClick={() => handleEquip(item)}
                        disabled={isActing}
                        className="w-full py-2.5 px-4 rounded-full bg-[#1E50A0] hover:bg-[#153e7e] text-white font-bold text-xs flex items-center justify-center gap-1.5 shadow-sm transition-all cursor-pointer"
                      >
                        <Sparkles className="w-3.5 h-3.5" />
                        <span>Надеть предмет</span>
                      </button>
                    )
                  ) : (
                    <button
                      onClick={() => handleBuy(item)}
                      disabled={!canAfford || isActing}
                      className={`w-full py-2.5 px-4 rounded-full font-bold text-xs flex items-center justify-center gap-1.5 transition-all ${
                        canAfford
                          ? 'bg-[#0C2340] hover:bg-[#1A4480] text-white shadow-sm'
                          : 'bg-slate-100 text-slate-400 cursor-not-allowed'
                      }`}
                    >
                      <Coins className="w-3.5 h-3.5" />
                      <span>
                        {isActing
                          ? 'Покупка...'
                          : canAfford
                          ? `Купить за ${item.price} Ҡурай`
                          : `Не хватает ${item.price - currency} монет`}
                      </span>
                    </button>
                  )}
                </div>
              </div>
            );
          })}
        </div>
      )}

      {/* How to earn Kurai info card */}
      <div className="bg-[#F8FAFC] rounded-3xl p-6 border border-slate-200 flex flex-col sm:flex-row items-center justify-between gap-4">
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 rounded-2xl bg-amber-100 flex items-center justify-center text-amber-700">
            <Coins className="w-5 h-5" />
          </div>
          <div>
            <div className="text-xs font-bold text-slate-800">Как получить монеты Ҡурай?</div>
            <div className="text-[11px] text-slate-500">
              Выполняйте ежедневные задания (+5), проходите квесты (+10), поддерживайте стрик (+20) и повышайте уровень (+50)!
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};
