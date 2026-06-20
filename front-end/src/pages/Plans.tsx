import React from 'react';
import { TENANT_URL } from '../config';
import { useApi } from '../hooks/useApi';

const PLANS = [
  {
    key: 'starter',
    label: 'Starter',
    price: '$29',
    period: '/mo',
    tagline: 'Perfect for small teams getting started',
    features: [
      '1,000 email scans per month',
      '90-day audit log retention',
      'Up to 5 team members',
      'Gmail integration',
      'AI compliance checks',
    ],
    trial: true,
    cta: 'Start 14-Day Free Trial',
    highlight: true,
  },
  {
    key: 'pro',
    label: 'Pro',
    price: '$99',
    period: '/mo',
    tagline: 'For growing teams with higher volume',
    features: [
      '10,000 email scans per month',
      '1-year audit log retention',
      'Up to 25 team members',
      'Gmail integration',
      'Priority AI processing',
    ],
    trial: false,
    cta: 'Get Started',
    highlight: false,
  },
  {
    key: 'business',
    label: 'Business',
    price: '$299',
    period: '/mo',
    tagline: 'Unlimited scale for large organizations',
    features: [
      'Unlimited email scans',
      '3-year audit log retention',
      'Unlimited team members',
      'Gmail integration',
      'Dedicated support',
    ],
    trial: false,
    cta: 'Get Started',
    highlight: false,
  },
];

interface PlansProps {
  onGoToSettings?: () => void;
}

export function Plans({ onGoToSettings }: PlansProps) {
  const { apiFetch } = useApi();

  const startCheckout = async (plan: string) => {
    try {
      const res = await apiFetch(`${TENANT_URL}/v1/billing/checkout`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ plan }),
      });
      const body = await res.json();
      if (body.data?.checkout_url) window.location.href = body.data.checkout_url;
    } catch { /* ignore */ }
  };

  return (
    <div className="min-h-screen bg-gray-50 py-12 px-4">
      <div className="max-w-4xl mx-auto">
        <div className="text-center mb-10">
          <h1 className="text-2xl font-bold text-gray-900 mb-2">Choose Your Plan</h1>
          <p className="text-gray-500 text-sm">
            Start protecting your team's email today. Cancel anytime.
          </p>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-5">
          {PLANS.map((plan) => (
            <div
              key={plan.key}
              className={`bg-white rounded-2xl border flex flex-col ${
                plan.highlight
                  ? 'border-brand shadow-lg ring-1 ring-brand/20'
                  : 'border-gray-100 shadow-sm'
              }`}
              style={{ padding: '24px' }}
            >
              {plan.highlight && (
                <div className="mb-3">
                  <span
                    className="text-[10px] font-semibold uppercase tracking-wider px-2.5 py-1 rounded-full text-white"
                    style={{ background: '#3d9970' }}
                  >
                    Most Popular
                  </span>
                </div>
              )}
              <h2 className="text-lg font-bold text-gray-900">{plan.label}</h2>
              <p className="text-xs text-gray-400 mt-0.5 mb-4">{plan.tagline}</p>
              <div className="flex items-end gap-0.5 mb-6">
                <span className="text-3xl font-bold text-gray-900">{plan.price}</span>
                <span className="text-sm text-gray-400 mb-1">{plan.period}</span>
              </div>
              <ul className="space-y-2.5 flex-1 mb-6">
                {plan.features.map((f) => (
                  <li key={f} className="flex items-start gap-2 text-sm text-gray-600">
                    <svg
                      className="flex-shrink-0 mt-0.5"
                      width="16"
                      height="16"
                      viewBox="0 0 20 20"
                      fill="#3d9970"
                    >
                      <path
                        fillRule="evenodd"
                        d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z"
                        clipRule="evenodd"
                      />
                    </svg>
                    {f}
                  </li>
                ))}
              </ul>
              <button
                onClick={() => startCheckout(plan.key)}
                className={`w-full py-2.5 rounded-xl text-sm font-semibold transition-colors ${
                  plan.highlight
                    ? 'text-white hover:opacity-90'
                    : 'border border-gray-200 text-gray-700 hover:bg-gray-50'
                }`}
                style={plan.highlight ? { background: '#3d9970' } : {}}
              >
                {plan.cta}
              </button>
              {plan.trial && (
                <p className="text-xs text-center text-gray-400 mt-2">
                  14 days free · Card required · Cancel anytime
                </p>
              )}
            </div>
          ))}
        </div>

        {onGoToSettings && (
          <p className="text-center text-xs text-gray-400 mt-8">
            Already on a plan?{' '}
            <button
              onClick={onGoToSettings}
              className="bg-transparent border-none cursor-pointer p-0 text-xs"
              style={{ color: '#3d9970' }}
            >
              Manage billing in Settings
            </button>
          </p>
        )}
      </div>
    </div>
  );
}
