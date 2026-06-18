import React, { useEffect, useState, useCallback } from 'react';
import { TENANT_URL } from '../config';
import { useApi } from '../hooks/useApi';
import { useAuth } from '../context/AuthContext';
import { useToast } from '../context/ToastContext';
import { QuarantineItem } from '../types';
import { StatusPill, PriorityPill } from '../components/Badge';
import { fmtTimeShort } from '../utils/format';

interface QuarantineProps {
  onBadgeChange: (count: number) => void;
}

export function Quarantine({ onBadgeChange }: QuarantineProps) {
  const { apiFetch } = useApi();
  const { role } = useAuth();
  const { toast } = useToast();
  const [items, setItems] = useState<QuarantineItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [reviewNote, setReviewNote] = useState('');
  const [reviewAlert, setReviewAlert] = useState<{ ok: boolean; msg: string } | null>(null);
  const [reviewLoading, setReviewLoading] = useState(false);

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const res = await apiFetch(`${TENANT_URL}/v1/quarantine?status=`);
      const data = await res.json();
      const list: QuarantineItem[] = data.data || [];
      setItems(list);
      const pending = list.filter((e) => e.status === 'pending').length;
      onBadgeChange(pending);
    } finally {
      setLoading(false);
    }
  }, [apiFetch, onBadgeChange]);

  useEffect(() => {
    load();
  }, [load]);

  const selected = items.find((i) => i.id === selectedId) || null;

  const doReview = async (id: string, action: 'release' | 'reject') => {
    const res = await apiFetch(`${TENANT_URL}/v1/quarantine/${id}/review`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ action }),
    });
    if (res.ok) {
      const newStatus = action === 'release' ? 'released' : 'rejected';
      const updated = items.map((e) => (e.id === id ? { ...e, status: newStatus as QuarantineItem['status'] } : e));
      setItems(updated);
      const pending = updated.filter((e) => e.status === 'pending').length;
      onBadgeChange(pending);
      toast(action === 'release' ? 'Email released and delivered.' : 'Email permanently erased.');
    } else {
      const d = await res.json();
      toast(d.message || 'Action failed.', 'error');
    }
  };

  const submitReviewRequest = async (quarantineId: string) => {
    setReviewLoading(true);
    setReviewAlert(null);
    try {
      const res = await apiFetch(`${TENANT_URL}/v1/quarantine/${quarantineId}/release-request`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ note: reviewNote }),
      });
      const data = await res.json();
      setReviewAlert({
        ok: res.ok,
        msg: res.ok ? 'Request sent to your account owner.' : (data.message || 'Failed.'),
      });
    } catch {
      setReviewAlert({ ok: false, msg: 'Network error.' });
    } finally {
      setReviewLoading(false);
    }
  };

  const scopeNote = role === 'owner' ? 'All team quarantine' : 'Your emails only';

  const renderDetail = () => {
    if (!selected) {
      return (
        <div className="flex-1 flex flex-col items-center justify-center gap-2 text-gray-300">
          <svg width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
            <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/>
          </svg>
          <p className="text-sm text-gray-400 m-0">Select an email to review</p>
        </div>
      );
    }

    const isPending = selected.status === 'pending';
    const isHighNonOwner = selected.priority === 'high' && role !== 'owner';
    const canAct = isPending && !isHighNonOwner;

    return (
      <div className="flex-1 flex flex-col overflow-hidden">
        <div className="px-6 py-4 border-b border-gray-100 flex-shrink-0">
          <div className="flex items-start gap-2 mb-3">
            <div className="mt-0.5 flex-shrink-0">
              <PriorityPill priority={selected.priority} />
            </div>
            <h3 className="text-base font-semibold text-gray-900 m-0 flex-1 leading-snug">
              {selected.subject || '(no subject)'}
            </h3>
          </div>
          <div className="text-xs text-gray-500 space-y-0.5">
            <div><span className="font-medium text-gray-600">From:</span> {selected.email_from}</div>
            <div><span className="font-medium text-gray-600">To:</span> {selected.email_to}</div>
          </div>
        </div>
        <div className="flex-1 overflow-y-auto px-6 py-4">
          <div className="flex flex-wrap gap-1 mb-3">
            {(selected.violations || []).map((v, i) => (
              <span key={i} className="bg-red-100 text-red-700 text-xs px-2.5 py-0.5 rounded-full">{v}</span>
            ))}
          </div>
          {selected.reasoning && (
            <div className="text-xs text-gray-500 bg-amber-50 border border-amber-100 rounded-lg px-3 py-2.5 mb-4 leading-relaxed">
              {selected.reasoning}
            </div>
          )}
          <pre className="text-sm text-gray-600 whitespace-pre-wrap font-sans leading-relaxed bg-gray-50 border border-gray-100 rounded-lg p-4 max-h-72 overflow-y-auto m-0">
            {selected.body || '(no body)'}
          </pre>
          {isPending && isHighNonOwner && (
            <div className="mt-4 bg-gray-50 border border-gray-200 rounded-xl p-4">
              <p className="text-sm font-semibold text-gray-800 mb-2">Request owner review</p>
              <textarea
                rows={2}
                placeholder="Why might this be a false positive?"
                className="w-full border border-gray-200 rounded-lg px-3 py-2 text-sm resize-none focus:outline-none focus:ring-2 focus:border-brand mb-2"
                value={reviewNote}
                onChange={(e) => setReviewNote(e.target.value)}
              />
              <button
                onClick={() => submitReviewRequest(selected.id)}
                disabled={reviewLoading}
                className="text-sm px-4 py-1.5 border border-brand text-brand rounded-lg hover:bg-brand/5 transition-colors font-medium disabled:opacity-70"
                style={{ borderColor: '#3d9970', color: '#3d9970' }}
              >
                Send to Owner
              </button>
              {reviewAlert && (
                <div className={`mt-2 text-sm px-3 py-2 rounded-lg ${reviewAlert.ok ? 'bg-green-50 text-green-700 border border-green-200' : 'bg-red-50 text-red-700 border border-red-200'}`}>
                  {reviewAlert.msg}
                </div>
              )}
            </div>
          )}
        </div>
        <div className="px-6 py-3.5 border-t border-gray-100 flex gap-2 items-center bg-white flex-shrink-0">
          {!isPending && (
            <span className="text-xs text-gray-400">
              This email has been <strong>{selected.status}</strong>.
            </span>
          )}
          {canAct && (
            <>
              <button
                onClick={() => doReview(selected.id, 'release')}
                className="bg-green-600 hover:bg-green-700 text-white text-sm font-medium px-4 py-2 rounded-lg transition-colors"
              >
                ✓ Release &amp; Deliver
              </button>
              <button
                onClick={() => doReview(selected.id, 'reject')}
                className="border border-red-200 text-red-600 text-sm font-medium px-4 py-2 rounded-lg hover:bg-red-50 transition-colors"
              >
                ✕ Permanently Erase
              </button>
            </>
          )}
        </div>
      </div>
    );
  };

  return (
    <div style={{ height: 'calc(100vh - 58px)', display: 'flex' }}>
      {/* List */}
      <div
        style={{
          width: '340px',
          flexShrink: 0,
          borderRight: '1px solid #f3f4f6',
          display: 'flex',
          flexDirection: 'column',
          overflow: 'hidden',
        }}
      >
        <div className="px-4 py-3 border-b border-gray-100 flex items-center gap-2 flex-shrink-0">
          <p className="text-xs text-gray-400 flex-1 m-0">{scopeNote}</p>
          <button onClick={load} className="p-1.5 rounded-lg text-gray-400 hover:bg-gray-100 transition-colors">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <polyline points="23 4 23 10 17 10"/>
              <path d="M20.49 15a9 9 0 1 1-.49-4.79"/>
            </svg>
          </button>
        </div>
        {loading && (
          <div className="px-4 py-8 text-center text-sm text-gray-400">Loading…</div>
        )}
        <div className="flex-1 overflow-y-auto">
          {!loading && items.length === 0 && (
            <div className="px-4 py-12 text-center text-sm text-gray-400">No quarantine items.</div>
          )}
          {items.map((e) => {
            const isActive = e.id === selectedId;
            return (
              <div
                key={e.id}
                onClick={() => { setSelectedId(e.id); setReviewNote(''); setReviewAlert(null); }}
                className={`flex items-start gap-2.5 px-4 py-3 border-b border-gray-50 cursor-pointer ${isActive ? 'bg-green-50' : ''} ${e.priority === 'high' && e.status === 'pending' && !isActive ? 'bg-red-50/30' : ''} ${e.priority !== 'high' && e.status === 'pending' && !isActive ? 'bg-amber-50/30' : ''}`}
              >
                <div className={`w-2 h-2 rounded-full flex-shrink-0 mt-1.5 ${e.priority === 'high' ? 'bg-red-400' : 'bg-amber-400'}`} />
                <div className="flex-1 min-w-0">
                  <div className={`text-[13px] truncate ${e.status === 'pending' ? 'font-semibold text-gray-800' : 'text-gray-500'}`}>
                    {e.email_from}
                  </div>
                  <div className="text-xs text-gray-400 truncate mt-0.5">{e.subject || '(no subject)'}</div>
                </div>
                <div className="flex flex-col items-end gap-1 flex-shrink-0">
                  <StatusPill status={e.status} />
                  <span className="text-[11px] text-gray-300">{fmtTimeShort(e.created_at)}</span>
                </div>
              </div>
            );
          })}
        </div>
      </div>

      {/* Detail */}
      <div style={{ flex: 1, display: 'flex', flexDirection: 'column', overflow: 'hidden' }}>
        {renderDetail()}
      </div>
    </div>
  );
}
