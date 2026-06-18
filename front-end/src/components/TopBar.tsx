import React from 'react';

interface TopBarProps {
  title: string;
  email: string;
  onToggleSidebar: () => void;
}

export function TopBar({ title, email, onToggleSidebar }: TopBarProps) {
  return (
    <div className="sticky top-0 z-[100] bg-white border-b border-gray-100 h-[58px] px-6 flex items-center gap-3">
      <button
        onClick={onToggleSidebar}
        className="p-1.5 rounded-lg text-gray-400 hover:bg-gray-100 hover:text-gray-700 transition-colors"
      >
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
          <line x1="3" y1="6" x2="21" y2="6"/>
          <line x1="3" y1="12" x2="21" y2="12"/>
          <line x1="3" y1="18" x2="21" y2="18"/>
        </svg>
      </button>
      <h1 className="text-[15px] font-semibold text-gray-900 flex-1 m-0">{title}</h1>
      <span className="text-xs text-gray-400">{email}</span>
    </div>
  );
}
