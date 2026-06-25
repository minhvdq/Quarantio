import React, { useEffect, useState, useCallback } from 'react';
import { TENANT_URL } from '../config';
import { useApi } from '../hooks/useApi';
import { useAuth } from '../context/AuthContext';
import { QuarantineItem, AuditEntry, Member } from '../types';
import { StatusPill } from '../components/Badge';
import { fmtTimeShort } from '../utils/format';

interface DashboardProps {
  onNavigateToQuarantine: () => void;
}

export function Dashboard({ onNavigateToQuarantine }: DashboardProps) {
  const { apiFetch } = useApi();
  const { role } = useAuth();
  const isOwner = role === 'owner';
  const [scansToday, setScansToday] = useState<number | null>(null);
  const [quarantineCount, setQuarantineCount] = useState<number | null>(null);
  const [pendingCount, setPendingCount] = useState<number | null>(null);
  const [memberCount, setMemberCount] = useState<number | null>(null);
  const [recent, setRecent] = useState<QuarantineItem[]>([]);
  const [loading, setLoading] = useState(true);

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const [qRes, auditRes] = await Promise.all([
        apiFetch(`${TENANT_URL}/v1/quarantine?status=`),
        apiFetch(`${TENANT_URL}/v1/audit`),
      ]);
      const qData: QuarantineItem[] = qRes.ok ? ((await qRes.json()).data || []) : [];
      const auditData: AuditEntry[] = auditRes.ok ? ((await auditRes.json()).data || []) : [];
      let members: Member[] = [];
      if (isOwner) {
        const membersRes = await apiFetch(`${TENANT_URL}/v1/members`);
        const membData = membersRes.ok ? await membersRes.json() : [];
        members = Array.isArray(membData) ? membData : (membData.data || []);
      }

      const pending = qData.filter((e) => e.status === 'pending').length;
      const todayStr = new Date().toLocaleDateString();
      const today = auditData.filter(
        (e) => new Date(e.created_at).toLocaleDateString() === todayStr
      ).length;

      setScansToday(today);
      setQuarantineCount(qData.length);
      setPendingCount(pending);
      setMemberCount(members.length);
      setRecent(qData.slice(0, 5));
    } finally {
      setLoading(false);
    }
  }, [apiFetch]);

  useEffect(() => {
    load();
  }, [load]);

  const kpiCards = [
    { label: 'Scans Today', value: scansToday, accent: 'border-t-brand' },
    { label: 'Quarantined', value: quarantineCount, accent: 'border-t-red-400' },
    { label: 'Pending Review', value: pendingCount, accent: 'border-t-amber-400' },
    ...(isOwner ? [{ label: 'Team Members', value: memberCount, accent: 'border-t-blue-400' }] : []),
  ];

  return (
    <div className="p-6">
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
        {kpiCards.map((k) => (
          <div
            key={k.label}
            className={`bg-white rounded-xl border border-gray-100 p-5 shadow-sm border-t-2 ${k.accent}`}
          >
            <div className="text-xs font-semibold uppercase tracking-wider text-gray-400 mb-2">
              {k.label}
            </div>
            <div className="text-3xl font-bold text-gray-900">
              {k.value === null ? '—' : k.value}
            </div>
          </div>
        ))}
      </div>

      <div className="bg-white rounded-xl border border-gray-100 shadow-sm overflow-hidden">
        <div className="px-5 py-4 border-b border-gray-100 flex items-center justify-between">
          <h2 className="text-sm font-semibold text-gray-800">Recent Quarantine</h2>
          <button
            onClick={onNavigateToQuarantine}
            className="text-xs hover:underline"
            style={{ color: '#3d9970' }}
          >
            View all
          </button>
        </div>
        <div className="divide-y divide-gray-50">
          {loading ? (
            <div className="px-5 py-8 text-center text-sm text-gray-400">Loading…</div>
          ) : recent.length === 0 ? (
            <div className="px-5 py-8 text-center text-sm text-gray-400">
              No quarantine items yet.
            </div>
          ) : (
            recent.map((e) => (
              <div
                key={e.id}
                className="px-5 py-3 flex items-center gap-3 hover:bg-gray-50 cursor-pointer"
                onClick={onNavigateToQuarantine}
              >
                <div
                  className={`w-2 h-2 rounded-full flex-shrink-0 ${
                    e.priority === 'high' ? 'bg-red-400' : 'bg-amber-400'
                  }`}
                />
                <div className="flex-1 min-w-0">
                  <div className="text-sm font-medium text-gray-800 truncate">
                    {e.email_from}
                  </div>
                  <div className="text-xs text-gray-400 truncate">
                    {e.subject || '(no subject)'}
                  </div>
                </div>
                <StatusPill status={e.status} />
                <div className="text-xs text-gray-300">{fmtTimeShort(e.created_at)}</div>
              </div>
            ))
          )}
        </div>
      </div>
    </div>
  );
}
