import React from 'react';
import { TENANT_URL } from '../config';

export function Landing() {
  return (
    <div style={{ fontFamily: '-apple-system,BlinkMacSystemFont,"Segoe UI",sans-serif', color: '#1a2e23', margin: 0 }}>
      <style>{`
        :root { --g50:#f0faf5; --g100:#e0f2ea; --g200:#b8e0ca; --g500:#3d9970; --g600:#2f7a58; --g700:#1e5c3f; }
        * { box-sizing:border-box; }
        .nav-bar { background:#fff; border-bottom:1px solid var(--g100); padding:0 24px; height:60px; display:flex; align-items:center; justify-content:space-between; position:sticky; top:0; z-index:100; }
        .nav-logo { display:flex; align-items:center; gap:10px; font-weight:700; font-size:1.1rem; color:var(--g700); text-decoration:none; }
        .nav-logo img { width:32px; height:32px; object-fit:contain; }
        .nav-actions { display:flex; align-items:center; gap:12px; }
        .btn-nav { padding:7px 20px; border-radius:8px; font-size:.875rem; font-weight:500; text-decoration:none; transition:all .15s; display:inline-block; }
        .btn-nav-ghost { color:var(--g600); border:1.5px solid var(--g200); background:transparent; }
        .btn-nav-ghost:hover { background:var(--g50); color:var(--g700); }
        .btn-nav-primary { background:var(--g500); color:#fff; border:1.5px solid var(--g500); }
        .btn-nav-primary:hover { background:var(--g600); border-color:var(--g600); color:#fff; }
        .hero { background:linear-gradient(135deg,var(--g700) 0%,var(--g500) 60%,#5cb88a 100%); padding:80px 24px 96px; text-align:center; }
        .hero-badge { display:inline-flex; align-items:center; gap:6px; background:rgba(255,255,255,.15); border:1px solid rgba(255,255,255,.25); border-radius:999px; padding:5px 14px; font-size:.8rem; color:rgba(255,255,255,.9); margin-bottom:24px; }
        .hero h1 { font-size:clamp(2rem,5vw,3.2rem); font-weight:800; color:#fff; line-height:1.15; letter-spacing:-.03em; margin-bottom:20px; }
        .hero p { font-size:1.1rem; color:rgba(255,255,255,.85); max-width:580px; margin:0 auto 36px; line-height:1.6; }
        .hero-actions { display:flex; gap:12px; justify-content:center; flex-wrap:wrap; }
        .btn-hero-primary { background:#fff; color:var(--g700); border:none; padding:14px 32px; border-radius:10px; font-size:.975rem; font-weight:700; text-decoration:none; transition:all .15s; display:inline-flex; align-items:center; gap:8px; }
        .btn-hero-primary:hover { background:var(--g50); color:var(--g700); transform:translateY(-1px); box-shadow:0 6px 20px rgba(0,0,0,.15); }
        .btn-hero-ghost { background:rgba(255,255,255,.15); color:#fff; border:1.5px solid rgba(255,255,255,.35); padding:14px 28px; border-radius:10px; font-size:.975rem; font-weight:500; text-decoration:none; transition:all .15s; }
        .btn-hero-ghost:hover { background:rgba(255,255,255,.22); color:#fff; }
        .social-proof { background:#fff; border-bottom:1px solid var(--g100); padding:18px 24px; text-align:center; }
        .social-proof p { color:#6b8c7a; font-size:.85rem; margin:0; }
        .section { padding:72px 24px; }
        .section-label { font-size:.8rem; font-weight:700; text-transform:uppercase; letter-spacing:.1em; color:var(--g500); margin-bottom:10px; }
        .section h2 { font-size:clamp(1.6rem,3.5vw,2.2rem); font-weight:800; color:var(--g700); letter-spacing:-.02em; margin-bottom:16px; }
        .section-sub { color:#6b8c7a; font-size:1rem; max-width:520px; line-height:1.6; }
        .feature-grid { display:grid; grid-template-columns:repeat(auto-fit,minmax(260px,1fr)); gap:24px; margin-top:48px; }
        .feature-card { background:#fff; border:1px solid var(--g100); border-radius:14px; padding:28px 24px; transition:box-shadow .2s,transform .2s; }
        .feature-card:hover { box-shadow:0 8px 24px rgba(61,153,112,.12); transform:translateY(-2px); }
        .feature-icon { width:44px; height:44px; border-radius:10px; background:var(--g50); border:1px solid var(--g100); display:flex; align-items:center; justify-content:center; font-size:1.3rem; margin-bottom:16px; }
        .feature-card h4 { font-size:1rem; font-weight:700; color:var(--g700); margin-bottom:8px; }
        .feature-card p { font-size:.875rem; color:#6b8c7a; line-height:1.6; margin:0; }
        .how-section { background:var(--g50); border-top:1px solid var(--g100); border-bottom:1px solid var(--g100); }
        .step { display:flex; gap:20px; align-items:flex-start; }
        .step-num { width:36px; height:36px; border-radius:50%; background:var(--g500); color:#fff; font-weight:700; font-size:.9rem; display:flex; align-items:center; justify-content:center; flex-shrink:0; margin-top:2px; }
        .steps-col { display:flex; flex-direction:column; gap:24px; }
        .pricing-grid { display:grid; grid-template-columns:repeat(auto-fit,minmax(220px,1fr)); gap:20px; margin-top:48px; }
        .plan-card { background:#fff; border:1.5px solid var(--g100); border-radius:14px; padding:28px 24px; position:relative; transition:box-shadow .2s; }
        .plan-card:hover { box-shadow:0 8px 24px rgba(61,153,112,.1); }
        .plan-card.popular { border-color:var(--g500); }
        .popular-badge { position:absolute; top:-13px; left:50%; transform:translateX(-50%); background:var(--g500); color:#fff; font-size:.72rem; font-weight:700; padding:3px 14px; border-radius:999px; white-space:nowrap; }
        .plan-name { font-weight:700; font-size:1rem; color:var(--g700); margin-bottom:4px; }
        .plan-price { font-size:2rem; font-weight:800; color:var(--g700); line-height:1; margin-bottom:4px; }
        .plan-price span { font-size:.95rem; font-weight:400; color:#6b8c7a; }
        .plan-desc { font-size:.82rem; color:#6b8c7a; margin-bottom:20px; }
        .plan-features { list-style:none; padding:0; margin:0; font-size:.85rem; color:#4a6858; }
        .plan-features li { padding:5px 0; display:flex; align-items:center; gap:8px; }
        .plan-features li::before { content:"✓"; color:var(--g500); font-weight:700; flex-shrink:0; }
        .cta-section { background:linear-gradient(135deg,var(--g700),var(--g500)); padding:72px 24px; text-align:center; }
        .cta-section h2 { font-size:clamp(1.6rem,3.5vw,2.2rem); font-weight:800; color:#fff; margin-bottom:16px; }
        .cta-section p { color:rgba(255,255,255,.85); font-size:1rem; margin-bottom:32px; }
        footer { background:#fff; border-top:1px solid var(--g100); padding:28px 24px; }
        .footer-inner { max-width:1100px; margin:0 auto; display:flex; align-items:center; justify-content:space-between; flex-wrap:wrap; gap:16px; }
        .footer-logo { font-weight:700; color:var(--g700); font-size:.95rem; display:flex; align-items:center; gap:8px; }
        .footer-logo img { width:24px; height:24px; object-fit:contain; }
        .footer-links { display:flex; gap:20px; }
        .footer-links a { color:#6b8c7a; font-size:.82rem; text-decoration:none; }
        .footer-links a:hover { color:var(--g600); }
        .text-center { text-align:center; }
        @media(max-width:600px){ .hero{padding:56px 20px 72px;} .btn-nav-ghost{display:none;} }
      `}</style>

      {/* Nav */}
      <nav className="nav-bar">
        <a href="/" className="nav-logo">
          <img src="/logo.png" alt="Quarantio" />
          Quarantio
        </a>
        <div className="nav-actions">
          <a href="/" className="btn-nav btn-nav-ghost">Sign In</a>
          <a href="/" className="btn-nav btn-nav-primary">Start Free Trial</a>
        </div>
      </nav>

      {/* Hero */}
      <section className="hero">
        <div className="hero-badge">✦ AI-powered email compliance</div>
        <h1>Guard your inbox.<br />Protect your team.</h1>
        <p>Quarantio scans every email for policy violations in real-time — automatically removing risky messages before they reach your inbox.</p>
        <div className="hero-actions">
          <a href={`${TENANT_URL}/auth/google/login`} className="btn-hero-primary">
            <GoogleColorIcon />
            Continue with Google
          </a>
          <a href="/#pricing" className="btn-hero-ghost">See pricing</a>
        </div>
      </section>

      {/* Social proof */}
      <div className="social-proof">
        <p>🛡️ Trusted by compliance-focused teams · No MX changes required · Works with any Gmail account</p>
      </div>

      {/* Features */}
      <section className="section" style={{ background: '#fff' }}>
        <div style={{ maxWidth: '1100px', margin: '0 auto' }}>
          <div className="text-center">
            <p className="section-label">Features</p>
            <h2>Everything your team needs to stay compliant</h2>
            <p className="section-sub" style={{ margin: '0 auto' }}>From real-time interception to detailed audit logs — Quarantio handles compliance so your team doesn't have to think about it.</p>
          </div>
          <div className="feature-grid">
            {[
              { icon: '🛡️', title: 'Auto Guard', desc: 'Emails are scanned the moment they arrive. Policy violations are removed from the inbox before the user even sees them.' },
              { icon: '🤖', title: 'AI Compliance Engine', desc: 'Powered by Mistral AI with your custom policy documents. Upload your rulebook and Quarantio learns what\'s allowed.' },
              { icon: '🔍', title: 'Quarantine & Review', desc: 'Flagged emails land in a secure quarantine. Admins review, release, or permanently erase with full audit trails.' },
              { icon: '👥', title: 'Team Management', desc: 'Invite your team with a single click. Role-based access keeps owners in control. Real-time visibility across all mailboxes.' },
              { icon: '📊', title: 'Audit Logs', desc: 'Every scanned email is logged with verdict, violations, and reasoning. Export anytime for compliance reporting.' },
              { icon: '⚡', title: 'Instant Setup', desc: 'Sign in with Google, connect your mailbox, start your trial. No MX record changes, no IT department required.' },
            ].map((f) => (
              <div key={f.title} className="feature-card">
                <div className="feature-icon">{f.icon}</div>
                <h4>{f.title}</h4>
                <p>{f.desc}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* How it works */}
      <section className="section how-section">
        <div style={{ maxWidth: '720px', margin: '0 auto' }}>
          <div className="text-center" style={{ marginBottom: '40px' }}>
            <p className="section-label">How it works</p>
            <h2>Up and running in 3 steps</h2>
          </div>
          <div className="steps-col">
            {[
              { n: 1, title: 'Sign in with Google', desc: 'Your account and organization are created automatically. No forms, no passwords.' },
              { n: 2, title: 'Start your free trial & connect Gmail', desc: 'Activate your 14-day trial and grant Gmail access. Quarantio begins scanning within seconds.' },
              { n: 3, title: 'Upload your compliance policies', desc: 'Drop in your policy documents. The AI learns your rules and flags violations automatically.' },
            ].map((s) => (
              <div key={s.n} className="step">
                <div className="step-num">{s.n}</div>
                <div>
                  <h5 style={{ fontWeight: 600, marginBottom: '4px', fontSize: '1rem' }}>{s.title}</h5>
                  <p style={{ color: '#6b8c7a', fontSize: '.875rem', margin: 0 }}>{s.desc}</p>
                </div>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* Pricing */}
      <section className="section" id="pricing" style={{ background: '#fff' }}>
        <div style={{ maxWidth: '1100px', margin: '0 auto' }}>
          <div className="text-center">
            <p className="section-label">Pricing</p>
            <h2>Simple, transparent pricing</h2>
            <p className="section-sub" style={{ margin: '0 auto' }}>Start with a 14-day free trial. No credit card required.</p>
          </div>
          <div className="pricing-grid">
            <div className="plan-card">
              <div className="plan-name">Starter</div>
              <div className="plan-price">$29<span>/mo</span></div>
              <div className="plan-desc">For small teams getting started</div>
              <ul className="plan-features">
                <li>1,000 scans / month</li>
                <li>2 connected mailboxes</li>
                <li>90-day audit retention</li>
                <li>Quarantine &amp; review</li>
                <li>Email notifications</li>
              </ul>
            </div>
            <div className="plan-card popular">
              <div className="popular-badge">Most Popular</div>
              <div className="plan-name">Pro</div>
              <div className="plan-price">$99<span>/mo</span></div>
              <div className="plan-desc">For growing compliance teams</div>
              <ul className="plan-features">
                <li>10,000 scans / month</li>
                <li>10 connected mailboxes</li>
                <li>1-year audit retention</li>
                <li>Everything in Starter</li>
                <li>Priority support</li>
              </ul>
            </div>
            <div className="plan-card">
              <div className="plan-name">Business</div>
              <div className="plan-price">$299<span>/mo</span></div>
              <div className="plan-desc">For large organizations</div>
              <ul className="plan-features">
                <li>Unlimited scans</li>
                <li>Unlimited mailboxes</li>
                <li>3-year audit retention</li>
                <li>Everything in Pro</li>
                <li>SLA &amp; dedicated support</li>
              </ul>
            </div>
          </div>
        </div>
      </section>

      {/* CTA */}
      <section className="cta-section">
        <div style={{ maxWidth: '600px', margin: '0 auto' }}>
          <h2>Ready to protect your team's inbox?</h2>
          <p>Start your 14-day free trial. No credit card required. Cancel anytime.</p>
          <a href={`${TENANT_URL}/auth/google/login`} className="btn-hero-primary" style={{ display: 'inline-flex' }}>
            <GoogleColorIcon />
            Get started with Google
          </a>
        </div>
      </section>

      {/* Footer */}
      <footer>
        <div className="footer-inner">
          <div className="footer-logo">
            <img src="/logo.png" alt="" />
            Quarantio
          </div>
          <div className="footer-links">
            <a href="/landing">Home</a>
            <a href="/landing#privacy">Privacy Policy</a>
            <a href="mailto:support@quarantio.com">Contact</a>
          </div>
          <p style={{ color: '#9cb8a8', fontSize: '.78rem', margin: 0 }}>© 2026 Quarantio. All rights reserved.</p>
        </div>
      </footer>
    </div>
  );
}

function GoogleColorIcon() {
  return (
    <svg width="18" height="18" viewBox="0 0 48 48">
      <path fill="#EA4335" d="M24 9.5c3.54 0 6.71 1.22 9.21 3.6l6.85-6.85C35.9 2.38 30.47 0 24 0 14.62 0 6.51 5.38 2.56 13.22l7.98 6.19C12.43 13.72 17.74 9.5 24 9.5z"/>
      <path fill="#4285F4" d="M46.98 24.55c0-1.57-.15-3.09-.38-4.55H24v9.02h12.94c-.58 2.96-2.26 5.48-4.78 7.18l7.73 6c4.51-4.18 7.09-10.36 7.09-17.65z"/>
      <path fill="#FBBC05" d="M10.53 28.59c-.48-1.45-.76-2.99-.76-4.59s.27-3.14.76-4.59l-7.98-6.19C.92 16.46 0 20.12 0 24c0 3.88.92 7.54 2.56 10.78l7.97-6.19z"/>
      <path fill="#34A853" d="M24 48c6.48 0 11.93-2.13 15.89-5.81l-7.73-6c-2.18 1.48-4.97 2.31-8.16 2.31-6.26 0-11.57-4.22-13.47-9.91l-7.98 6.19C6.51 42.62 14.62 48 24 48z"/>
    </svg>
  );
}
