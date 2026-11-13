import React, { useState } from 'react';
import { Volume2, VolumeX } from 'lucide-react';
import { speakBashkir } from '../../utils/speech';

interface AudioSpeakerButtonProps {
  text: string;
  className?: string;
  size?: 'sm' | 'md' | 'lg';
  showLabel?: boolean;
  label?: string;
}

export const AudioSpeakerButton: React.FC<AudioSpeakerButtonProps> = ({
  text,
  className = '',
  size = 'md',
  showLabel = false,
  label = 'Озвучить',
}) => {
  const [isPlaying, setIsPlaying] = useState(false);

  const handleSpeak = (e: React.MouseEvent) => {
    e.stopPropagation();
    if (isPlaying) return;

    speakBashkir(text, {
      onStart: () => setIsPlaying(true),
      onEnd: () => setIsPlaying(false),
      onError: () => setIsPlaying(false),
    });
  };

  const sizeClasses = {
    sm: 'w-7 h-7 text-xs',
    md: 'w-9 h-9 text-sm',
    lg: 'w-11 h-11 text-base',
  }[size];

  const iconSizes = {
    sm: 'w-3.5 h-3.5',
    md: 'w-4 h-4',
    lg: 'w-5 h-5',
  }[size];

  return (
    <button
      type="button"
      onClick={handleSpeak}
      title="Прослушать произношение"
      className={`inline-flex items-center justify-center gap-1.5 rounded-full transition-all cursor-pointer ${
        isPlaying
          ? 'bg-[#1E50A0] text-white shadow-md scale-105 ring-2 ring-[#1E50A0]/40'
          : 'bg-slate-100 hover:bg-[#1E50A0]/10 text-slate-700 hover:text-[#1E50A0]'
      } ${showLabel ? 'px-3 py-1.5' : sizeClasses} ${className}`}
    >
      {isPlaying ? (
        <div className="flex items-center gap-0.5">
          <span className="w-1 h-3 bg-white rounded-full animate-bounce [animation-delay:-0.3s]" />
          <span className="w-1 h-4 bg-white rounded-full animate-bounce [animation-delay:-0.15s]" />
          <span className="w-1 h-2 bg-white rounded-full animate-bounce" />
        </div>
      ) : (
        <Volume2 className={iconSizes} />
      )}
      {showLabel && (
        <span className="font-bold text-xs select-none">
          {isPlaying ? 'Звучит...' : label}
        </span>
      )}
    </button>
  );
};
