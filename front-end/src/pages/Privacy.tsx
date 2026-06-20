import React from 'react';

const EFFECTIVE_DATE = 'June 19, 2025';
const CONTACT_EMAIL = 'privacy@quarantio.com';

export function Privacy() {
  return (
    <div style={{ fontFamily: '-apple-system,BlinkMacSystemFont,"Segoe UI",sans-serif', color: '#1a2e23', margin: 0 }}>
      <style>{`
        :root { --g50:#f0faf5; --g100:#e0f2ea; --g200:#b8e0ca; --g500:#3d9970; --g600:#2f7a58; --g700:#1e5c3f; }
        * { box-sizing:border-box; }
        .pp-nav { background:#fff; border-bottom:1px solid var(--g100); padding:0 24px; height:60px; display:flex; align-items:center; justify-content:space-between; position:sticky; top:0; z-index:100; }
        .pp-logo { display:flex; align-items:center; gap:10px; font-weight:700; font-size:1.1rem; color:var(--g700); text-decoration:none; }
        .pp-logo img { width:28px; height:28px; object-fit:contain; }
        .pp-back { font-size:.875rem; color:var(--g600); text-decoration:none; font-weight:500; }
        .pp-back:hover { color:var(--g700); }
        .pp-hero { background:linear-gradient(135deg,var(--g700) 0%,var(--g500) 100%); padding:48px 24px; text-align:center; }
        .pp-hero h1 { font-size:clamp(1.6rem,4vw,2.4rem); font-weight:800; color:#fff; margin:0 0 8px; letter-spacing:-.02em; }
        .pp-hero p { font-size:.9rem; color:rgba(255,255,255,.75); margin:0; }
        .pp-body { max-width:760px; margin:0 auto; padding:48px 24px 80px; }
        .pp-toc { background:var(--g50); border:1px solid var(--g100); border-radius:12px; padding:20px 24px; margin-bottom:40px; }
        .pp-toc h3 { font-size:.8rem; font-weight:700; text-transform:uppercase; letter-spacing:.08em; color:var(--g500); margin:0 0 12px; }
        .pp-toc ol { margin:0; padding-left:18px; }
        .pp-toc li { margin-bottom:6px; }
        .pp-toc a { font-size:.875rem; color:var(--g600); text-decoration:none; }
        .pp-toc a:hover { text-decoration:underline; }
        .pp-section { margin-bottom:40px; }
        .pp-section h2 { font-size:1.15rem; font-weight:700; color:var(--g700); border-bottom:1px solid var(--g100); padding-bottom:8px; margin-bottom:16px; }
        .pp-section p, .pp-section li { font-size:.9375rem; color:#3d5249; line-height:1.75; }
        .pp-section ul, .pp-section ol { padding-left:20px; margin:8px 0 16px; }
        .pp-section li { margin-bottom:6px; }
        .pp-section strong { color:#1a2e23; }
        .pp-highlight { background:var(--g50); border-left:3px solid var(--g500); border-radius:0 8px 8px 0; padding:12px 16px; margin:16px 0; font-size:.9rem; color:var(--g700); }
        .pp-table { width:100%; border-collapse:collapse; margin:12px 0 20px; font-size:.875rem; }
        .pp-table th { background:var(--g50); color:var(--g700); font-weight:600; text-align:left; padding:10px 14px; border:1px solid var(--g100); }
        .pp-table td { padding:10px 14px; border:1px solid var(--g100); color:#3d5249; vertical-align:top; }
        .pp-table tr:nth-child(even) td { background:#fafafa; }
        .pp-footer { border-top:1px solid var(--g100); margin-top:48px; padding-top:24px; font-size:.825rem; color:#6b8c7a; }
      `}</style>

      {/* Nav */}
      <nav className="pp-nav">
        <a href="/landing" className="pp-logo">
          <img src="/logo.png" alt="Quarantio" />
          Quarantio
        </a>
        <a href="/landing" className="pp-back">← Back to home</a>
      </nav>

      {/* Hero */}
      <div className="pp-hero">
        <h1>Privacy Policy</h1>
        <p>Effective date: {EFFECTIVE_DATE}</p>
      </div>

      {/* Body */}
      <div className="pp-body">

        {/* ToC */}
        <div className="pp-toc">
          <h3>Contents</h3>
          <ol>
            <li><a href="#who-we-are">Who We Are</a></li>
            <li><a href="#data-we-collect">Data We Collect</a></li>
            <li><a href="#google-data">Google User Data</a></li>
            <li><a href="#how-we-use">How We Use Your Data</a></li>
            <li><a href="#third-parties">Third-Party Services</a></li>
            <li><a href="#retention">Data Retention</a></li>
            <li><a href="#security">Security</a></li>
            <li><a href="#your-rights">Your Rights</a></li>
            <li><a href="#children">Children's Privacy</a></li>
            <li><a href="#changes">Changes to This Policy</a></li>
            <li><a href="#contact">Contact Us</a></li>
          </ol>
        </div>

        {/* 1 */}
        <div className="pp-section" id="who-we-are">
          <h2>1. Who We Are</h2>
          <p>
            Quarantio ("we", "us", or "our") is an AI-powered email compliance platform that helps
            organizations scan, classify, and manage email communications for policy violations and
            security threats. This Privacy Policy explains how we collect, use, store, and protect
            information when you use our service at <strong>quarantio.com</strong>.
          </p>
        </div>

        {/* 2 */}
        <div className="pp-section" id="data-we-collect">
          <h2>2. Data We Collect</h2>

          <p><strong>Account information</strong></p>
          <ul>
            <li>Email address, first name, and last name — provided via Google OAuth at sign-in</li>
            <li>Organization name and domain</li>
            <li>Role within your organization (Owner or Member)</li>
          </ul>

          <p><strong>Email content (when Gmail is connected)</strong></p>
          <ul>
            <li>Email sender, recipient, subject line, and body — fetched via the Gmail API for compliance scanning</li>
            <li>Emails flagged as non-compliant are stored in quarantine for your organization's review</li>
            <li>Emails classified as clean are never stored — only the scan result (verdict and metadata) is logged</li>
          </ul>

          <p><strong>Usage and audit data</strong></p>
          <ul>
            <li>Scan verdicts (CLEAN, LOW, MEDIUM, HIGH), violation categories, and AI reasoning</li>
            <li>Actions taken on quarantined emails (release, reject)</li>
            <li>Scan counts for billing enforcement</li>
          </ul>

          <p><strong>Billing information</strong></p>
          <ul>
            <li>Subscription plan and status — stored by us</li>
            <li>Payment card details — collected and stored exclusively by Stripe; we never see or store card numbers</li>
          </ul>

          <p><strong>Technical data</strong></p>
          <ul>
            <li>OAuth tokens for Gmail API access (encrypted at rest)</li>
            <li>Session tokens (short-lived, stored server-side)</li>
            <li>Standard server logs (IP address, request timestamps) for security and debugging</li>
          </ul>
        </div>

        {/* 3 */}
        <div className="pp-section" id="google-data">
          <h2>3. Google User Data</h2>

          <div className="pp-highlight">
            Quarantio's use and transfer of information received from Google APIs adheres to the{' '}
            <a href="https://developers.google.com/terms/api-services-user-data-policy" target="_blank" rel="noopener noreferrer">
              Google API Services User Data Policy
            </a>
            , including the Limited Use requirements.
          </div>

          <p><strong>What we access</strong></p>
          <p>
            When you connect your Gmail account, Quarantio requests the{' '}
            <code>https://www.googleapis.com/auth/gmail.readonly</code> scope. This allows us to read
            your email messages for the sole purpose of scanning them for compliance policy violations.
          </p>

          <p><strong>How we use Google data</strong></p>
          <ul>
            <li>We read emails to perform AI-powered compliance analysis</li>
            <li>Emails that pass the scan (CLEAN verdict) are analyzed and immediately discarded — we do not store clean email bodies</li>
            <li>Emails flagged as non-compliant are stored in your organization's quarantine for authorized reviewer action</li>
            <li>Email content is sent to Mistral AI's API for analysis and is subject to Mistral's privacy policy</li>
          </ul>

          <p><strong>What we never do with Google data</strong></p>
          <ul>
            <li>We do not sell, rent, or share Google user data with any third party for advertising or marketing</li>
            <li>We do not use Google data to serve advertisements</li>
            <li>We do not allow humans to read your email content except when you explicitly request support and grant access</li>
            <li>We do not use Google data for any purpose other than providing the compliance scanning service you signed up for</li>
          </ul>

          <p><strong>Revoking access</strong></p>
          <p>
            You can disconnect Gmail at any time from the Settings page. This immediately revokes our
            OAuth token and stops all email scanning. You can also revoke access directly from your
            Google Account at{' '}
            <a href="https://myaccount.google.com/permissions" target="_blank" rel="noopener noreferrer">
              myaccount.google.com/permissions
            </a>.
          </p>
        </div>

        {/* 4 */}
        <div className="pp-section" id="how-we-use">
          <h2>4. How We Use Your Data</h2>
          <ul>
            <li><strong>Provide the service</strong> — scan emails, enforce compliance policies, generate audit logs, and manage quarantined messages</li>
            <li><strong>Account management</strong> — authenticate users, manage organization membership and invitations</li>
            <li><strong>Billing</strong> — track scan usage against your plan limits and process subscription payments via Stripe</li>
            <li><strong>Security</strong> — detect abuse, prevent unauthorized access, and maintain service integrity</li>
            <li><strong>Communication</strong> — send transactional emails (email verification, invite notifications, billing receipts)</li>
            <li><strong>Service improvement</strong> — aggregate, anonymized usage metrics to improve scan accuracy and performance</li>
          </ul>
          <p>We do not use your data for behavioral advertising or sell it to data brokers.</p>
        </div>

        {/* 5 */}
        <div className="pp-section" id="third-parties">
          <h2>5. Third-Party Services</h2>
          <p>We share data with the following sub-processors only to the extent necessary to provide the service:</p>
          <table className="pp-table">
            <thead>
              <tr>
                <th>Provider</th>
                <th>Purpose</th>
                <th>Data shared</th>
              </tr>
            </thead>
            <tbody>
              <tr>
                <td><strong>Google LLC</strong></td>
                <td>OAuth authentication, Gmail API</td>
                <td>OAuth tokens; email content read via API</td>
              </tr>
              <tr>
                <td><strong>Mistral AI</strong></td>
                <td>AI compliance analysis</td>
                <td>Email subject and body for each scanned message</td>
              </tr>
              <tr>
                <td><strong>Stripe, Inc.</strong></td>
                <td>Payment processing</td>
                <td>Organization name; billing email; subscription info</td>
              </tr>
              <tr>
                <td><strong>Brevo (Sendinblue)</strong></td>
                <td>Transactional email delivery</td>
                <td>Recipient email address and message content</td>
              </tr>
            </tbody>
          </table>
          <p>
            We do not sell personal data. We require all sub-processors to maintain appropriate
            technical and organizational security measures.
          </p>
        </div>

        {/* 6 */}
        <div className="pp-section" id="retention">
          <h2>6. Data Retention</h2>
          <p>
            Audit logs, quarantine records, and policy embeddings are retained according to your
            organization's subscription plan:
          </p>
          <table className="pp-table">
            <thead>
              <tr>
                <th>Plan</th>
                <th>Audit log retention</th>
              </tr>
            </thead>
            <tbody>
              <tr><td>Starter</td><td>90 days</td></tr>
              <tr><td>Pro</td><td>1 year</td></tr>
              <tr><td>Business</td><td>3 years</td></tr>
            </tbody>
          </table>
          <p>
            Account information is retained for as long as your account exists. When an organization
            deletes its data or closes its account, all associated email content, audit logs, and
            embeddings are permanently deleted within 30 days. OAuth tokens are deleted immediately
            upon Gmail disconnection.
          </p>
        </div>

        {/* 7 */}
        <div className="pp-section" id="security">
          <h2>7. Security</h2>
          <p>We protect your data with the following measures:</p>
          <ul>
            <li>OAuth tokens are encrypted at rest using AES-256</li>
            <li>All data in transit is encrypted via TLS 1.2+</li>
            <li>Access to production systems is restricted to authorized personnel</li>
            <li>API keys and session tokens are hashed (SHA-256) before storage</li>
            <li>Stripe handles all payment card data — we are never in scope for PCI DSS</li>
          </ul>
          <p>
            No security system is perfect. If you discover a vulnerability, please report it to{' '}
            <a href={`mailto:${CONTACT_EMAIL}`}>{CONTACT_EMAIL}</a>.
          </p>
        </div>

        {/* 8 */}
        <div className="pp-section" id="your-rights">
          <h2>8. Your Rights</h2>
          <p>
            Depending on your jurisdiction (including the EU/EEA under GDPR and California under
            CCPA), you have the right to:
          </p>
          <ul>
            <li><strong>Access</strong> — request a copy of the personal data we hold about you</li>
            <li><strong>Export</strong> — download all your organization's data in JSON format from Settings → Export My Data</li>
            <li><strong>Delete</strong> — erase all audit logs, quarantine records, and policy embeddings from Settings → Erase All Data. To delete your account entirely, contact us.</li>
            <li><strong>Disconnect</strong> — revoke Gmail access at any time from Settings → Your Mailbox</li>
            <li><strong>Portability</strong> — your exported data is in a standard JSON format</li>
            <li><strong>Object</strong> — opt out of any processing not strictly necessary for the service</li>
          </ul>
          <p>
            To exercise any right not covered by the in-app controls, email{' '}
            <a href={`mailto:${CONTACT_EMAIL}`}>{CONTACT_EMAIL}</a>. We will respond within 30 days.
          </p>
        </div>

        {/* 9 */}
        <div className="pp-section" id="children">
          <h2>9. Children's Privacy</h2>
          <p>
            Quarantio is a business-to-business service intended for professional use. We do not
            knowingly collect personal data from anyone under the age of 16. If you believe we have
            inadvertently collected such data, please contact us and we will delete it promptly.
          </p>
        </div>

        {/* 10 */}
        <div className="pp-section" id="changes">
          <h2>10. Changes to This Policy</h2>
          <p>
            We may update this Privacy Policy from time to time. When we do, we will update the
            effective date at the top of this page and, for material changes, notify organization
            owners by email at least 14 days before the change takes effect. Continued use of the
            service after the effective date constitutes acceptance of the updated policy.
          </p>
        </div>

        {/* 11 */}
        <div className="pp-section" id="contact">
          <h2>11. Contact Us</h2>
          <p>
            If you have questions, concerns, or requests regarding this Privacy Policy or your
            personal data, please contact us at:
          </p>
          <p>
            <strong>Email:</strong>{' '}
            <a href={`mailto:${CONTACT_EMAIL}`}>{CONTACT_EMAIL}</a>
          </p>
          <p>
            For EU/EEA users, if you are not satisfied with our response you have the right to lodge
            a complaint with your local data protection authority.
          </p>
        </div>

        <div className="pp-footer">
          <p>© {new Date().getFullYear()} Quarantio. All rights reserved.</p>
        </div>
      </div>
    </div>
  );
}
