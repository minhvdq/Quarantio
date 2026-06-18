import React, { useState } from 'react';
import { BROKER_URL } from '../config';
import { useAuth } from '../context/AuthContext';

type AlertState = { type: 'success' | 'error' | 'warn'; msg: string } | null;

export function SendEmail() {
  const { tenantId } = useAuth();
  const [from, setFrom] = useState('');
  const [to, setTo] = useState('');
  const [subject, setSubject] = useState('');
  const [message, setMessage] = useState('');
  const [loading, setLoading] = useState(false);
  const [alert, setAlert] = useState<AlertState>(null);

  const inputCls =
    'w-full border border-gray-200 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:border-brand';

  const alertCls = {
    success: 'bg-green-50 text-green-700 border border-green-200',
    error: 'bg-red-50 text-red-700 border border-red-200',
    warn: 'bg-amber-50 text-amber-700 border border-amber-200',
  };

  const handleSend = async () => {
    if (!from || !to || !subject || !message) {
      setAlert({ type: 'warn', msg: 'All fields are required.' });
      return;
    }
    setAlert(null);
    setLoading(true);
    try {
      const res = await fetch(`${BROKER_URL}/handle`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          Action: 'mail',
          mail: { from, to, subject, message, tenant_id: tenantId },
        }),
      });
      const data = await res.json();
      if (data.error) {
        setAlert({ type: 'error', msg: data.message });
      } else {
        setAlert({ type: 'success', msg: 'Email submitted for compliance review.' });
        setFrom(''); setTo(''); setSubject(''); setMessage('');
      }
    } catch (e: unknown) {
      setAlert({ type: 'error', msg: 'Network error: ' + (e instanceof Error ? e.message : String(e)) });
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="p-6 max-w-2xl">
      <div className="mb-5">
        <h2 className="text-base font-semibold text-gray-900">Send Email</h2>
        <p className="text-sm text-gray-400 mt-0.5">Send through compliance review before delivery.</p>
      </div>
      {alert && (
        <div className={`mb-3 text-sm px-4 py-2.5 rounded-lg ${alertCls[alert.type]}`}>
          {alert.msg}
        </div>
      )}
      <div className="bg-white rounded-xl border border-gray-100 shadow-sm p-5">
        <div className="grid gap-3">
          <div>
            <label className="block text-xs font-medium text-gray-600 mb-1">From</label>
            <input type="email" className={inputCls} placeholder="you@company.com" value={from} onChange={(e) => setFrom(e.target.value)} />
          </div>
          <div>
            <label className="block text-xs font-medium text-gray-600 mb-1">To</label>
            <input type="email" className={inputCls} placeholder="recipient@example.com" value={to} onChange={(e) => setTo(e.target.value)} />
          </div>
          <div>
            <label className="block text-xs font-medium text-gray-600 mb-1">Subject</label>
            <input type="text" className={inputCls} value={subject} onChange={(e) => setSubject(e.target.value)} />
          </div>
          <div>
            <label className="block text-xs font-medium text-gray-600 mb-1">Message</label>
            <textarea rows={5} className={`${inputCls} resize-none`} value={message} onChange={(e) => setMessage(e.target.value)} />
          </div>
        </div>
        <button
          onClick={handleSend}
          disabled={loading}
          className="mt-4 w-full bg-brand hover:bg-brand-dark text-white font-medium text-sm py-2.5 rounded-lg transition-colors flex items-center justify-center gap-2 disabled:opacity-70"
        >
          {loading && (
            <span className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
          )}
          Send for Compliance Review
        </button>
      </div>
    </div>
  );
}
