import React from 'react';

type Verdict = 'CLEAN' | 'LOW' | 'MEDIUM' | 'HIGH' | string;
type Status = 'pending' | 'released' | 'rejected' | 'approved' | 'denied' | string;
type Priority = 'high' | 'medium' | 'low' | string;

export function VerdictBadge({ verdict }: { verdict: Verdict }) {
  const classes: Record<string, string> = {
    CLEAN: 'bg-green-100 text-green-700',
    LOW: 'bg-blue-100 text-blue-700',
    MEDIUM: 'bg-amber-100 text-amber-700',
    HIGH: 'bg-red-100 text-red-700',
  };
  const cls = classes[verdict] || 'bg-gray-100 text-gray-600';
  return (
    <span className={`text-xs px-2 py-0.5 rounded-full font-semibold ${cls}`}>
      {verdict}
    </span>
  );
}

export function StatusPill({ status }: { status: Status }) {
  const classes: Record<string, string> = {
    pending: 'bg-amber-100 text-amber-700',
    released: 'bg-green-100 text-green-700',
    rejected: 'bg-gray-100 text-gray-500',
    approved: 'bg-green-100 text-green-700',
    denied: 'bg-gray-100 text-gray-500',
  };
  const cls = classes[status] || 'bg-gray-100 text-gray-500';
  return (
    <span className={`text-[11px] px-2 py-0.5 rounded-full font-semibold ${cls}`}>
      {status}
    </span>
  );
}

export function PriorityPill({ priority }: { priority: Priority }) {
  if (priority === 'high') {
    return (
      <span className="text-xs px-2 py-0.5 rounded-full font-semibold bg-red-100 text-red-700">
        HIGH
      </span>
    );
  }
  return (
    <span className="text-xs px-2 py-0.5 rounded-full font-semibold bg-amber-100 text-amber-700">
      MEDIUM
    </span>
  );
}
