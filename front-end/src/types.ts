export interface QuarantineItem {
  id: string;
  email_from: string;
  email_to: string;
  subject: string;
  body: string;
  priority: 'high' | 'medium' | 'low';
  status: 'pending' | 'released' | 'rejected';
  violations: string[];
  reasoning: string;
  created_at: string;
}

export interface AuditEntry {
  id: string;
  email_from: string;
  email_subject: string;
  verdict: 'CLEAN' | 'LOW' | 'MEDIUM' | 'HIGH';
  action_taken: string;
  violations: string[];
  created_at: string;
}

export interface Member {
  id: string;
  email: string;
  first_name?: string;
  last_name?: string;
  role: 'owner' | 'user';
}

export interface Invite {
  id: string;
  email: string;
  inviter_email: string;
  created_at: string;
}

export interface Policy {
  filename: string;
  chunk_count: number;
  uploaded_at: string;
}

export interface Settings {
  auto_deliver_low: boolean;
  retention_days: number;
}

export interface BillingStatus {
  plan: 'free' | 'trial' | 'starter' | 'pro' | 'business';
  trial_ends_at?: string;
  scans_used?: number;
  scans_limit?: number;
}

export interface CheckResult {
  verdict: 'CLEAN' | 'LOW' | 'MEDIUM' | 'HIGH';
  violations: string[];
  reasoning: string;
  remediated_body?: string;
}

export interface ReleaseRequest {
  id: string;
  email_from: string;
  subject: string;
  requester_email: string;
  note?: string;
  status: 'pending' | 'approved' | 'denied';
  created_at: string;
}

export interface GmailStatus {
  connected: boolean;
  gmail_address?: string;
  last_scanned_at?: string;
}
