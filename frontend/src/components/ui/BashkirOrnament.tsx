import React from 'react';

export const BashkirOrnament: React.FC<{ className?: string }> = ({ className = "w-8 h-8" }) => (
  <svg
    viewBox="0 0 100 100"
    fill="currentColor"
    className={className}
    xmlns="http://www.w3.org/2000/svg"
  >
    {/* Traditional Bashkir Kuskar / Tamga geometric ornament */}
    <rect x="42" y="42" width="16" height="16" rx="2" transform="rotate(45 50 50)" />
    <path
      d="M50 15 L50 32 M50 68 L50 85 M15 50 L32 50 M68 50 L85 50"
      stroke="currentColor"
      strokeWidth="6"
      strokeLinecap="round"
    />
    {/* Corner hooks / spirals */}
    <path
      d="M36 28 C28 20, 20 28, 28 36 L38 46"
      fill="none"
      stroke="currentColor"
      strokeWidth="5"
      strokeLinecap="round"
    />
    <path
      d="M64 28 C72 20, 80 28, 72 36 L62 46"
      fill="none"
      stroke="currentColor"
      strokeWidth="5"
      strokeLinecap="round"
    />
    <path
      d="M36 72 C28 80, 20 72, 28 64 L38 54"
      fill="none"
      stroke="currentColor"
      strokeWidth="5"
      strokeLinecap="round"
    />
    <path
      d="M64 72 C72 80, 80 72, 72 64 L62 54"
      fill="none"
      stroke="currentColor"
      strokeWidth="5"
      strokeLinecap="round"
    />
  </svg>
);
