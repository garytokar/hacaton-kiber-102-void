import React from 'react';
import defaultAvatar from '../../assets/figma/avatar_user.png';
import { User } from '../../types';
import { Award, Sparkles, Sun } from 'lucide-react';

interface UserAvatarProps {
  user?: User | null;
  size?: 'xs' | 'sm' | 'md' | 'lg' | 'xl';
  className?: string;
  showBadge?: boolean;
}

export const UserAvatar: React.FC<UserAvatarProps> = ({
  user,
  size = 'md',
  className = '',
  showBadge = true,
}) => {
  const equippedFrame = user?.cosmetics?.equipped_frame;
  const equippedTitle = user?.cosmetics?.equipped_title;

  const sizeClasses = {
    xs: 'w-6 h-6',
    sm: 'w-8 h-8',
    md: 'w-10 h-10',
    lg: 'w-16 h-16',
    xl: 'w-24 h-24',
  }[size];

  // Specific cosmetic frame styling
  const frameStyles: Record<string, string> = {
    frame_gold: 'ring-2 ring-amber-400 ring-offset-2 ring-offset-white shadow-[0_0_14px_rgba(245,158,11,0.65)] border-amber-300',
    frame_sapphire: 'ring-2 ring-cyan-400 ring-offset-2 ring-offset-white shadow-[0_0_14px_rgba(6,182,212,0.65)] border-cyan-400',
    frame_emerald: 'ring-2 ring-emerald-400 ring-offset-2 ring-offset-white shadow-[0_0_14px_rgba(16,185,129,0.65)] border-emerald-400',
  };

  const activeFrameClass = equippedFrame && frameStyles[equippedFrame] ? frameStyles[equippedFrame] : 'border border-slate-200';

  const badgeSize = {
    xs: 'w-2.5 h-2.5 text-[8px]',
    sm: 'w-3.5 h-3.5 text-[9px]',
    md: 'w-4 h-4 text-[10px]',
    lg: 'w-6 h-6 text-xs',
    xl: 'w-8 h-8 text-sm',
  }[size];

  return (
    <div className={`relative inline-block select-none ${className}`}>
      <div
        className={`${sizeClasses} rounded-full overflow-hidden transition-all duration-300 ${activeFrameClass} bg-slate-100 flex items-center justify-center`}
      >
        <img
          src={defaultAvatar}
          alt={user?.full_name || user?.username || 'Хранитель'}
          className="w-full h-full object-cover"
          onError={(e) => {
            (e.target as HTMLElement).style.display = 'none';
          }}
        />
      </div>

      {/* Decorative solar rays for gold frame */}
      {equippedFrame === 'frame_gold' && (
        <span className="absolute -top-1 -right-1 text-amber-500 animate-pulse">
          <Sun className="w-3.5 h-3.5 drop-shadow" />
        </span>
      )}

      {/* Aksakal title badge */}
      {showBadge && equippedTitle === 'title_aksakal' && (
        <div
          title="Титул: Аксакал языка"
          className={`absolute -bottom-1 -right-1 ${badgeSize} bg-gradient-to-r from-amber-500 to-yellow-400 rounded-full flex items-center justify-center text-white ring-2 ring-white shadow-md`}
        >
          <Award className="w-full h-full p-0.5" />
        </div>
      )}
    </div>
  );
};
