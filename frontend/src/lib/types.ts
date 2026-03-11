// ---------------------------------------------------------------------------
// Shared TypeScript types mirroring the backend Go domain models.
// ---------------------------------------------------------------------------

// ---- Account (Chart of Accounts) ----

export type AccountType = "asset" | "liability" | "equity" | "revenue" | "expense";

export interface Account {
  id: string;
  code: string;
  name: string;
  type: AccountType;
  parent_id: string | null;
  level: number;
  is_system: boolean;
  is_control: boolean;
  is_active: boolean;
  balance: number;
  created_by?: string;
  created_at: string;
  updated_at: string;
}

export interface AccountTreeNode {
  account: Account;
  children?: AccountTreeNode[];
}

export interface CreateAccountRequest {
  code: string;
  name: string;
  parent_id: string;
  starting_balance?: number;
}

export interface UpdateAccountRequest {
  code?: string;
  name?: string;
  balance?: number;
}

// ---- Journal (Jurnal Umum) ----

export type JournalStatus = "draft" | "posted" | "reversed";

export interface JournalLine {
  id: string;
  journal_entry_id: string;
  account_id: string;
  account_code?: string;
  account_name?: string;
  description?: string;
  debit: number;
  credit: number;
  line_order: number;
  created_at: string;
}

export interface JournalEntry {
  id: string;
  entry_number: string;
  date: string;
  description: string;
  status: JournalStatus;
  total_debit: number;
  total_credit: number;
  reversal_of?: string;
  reversal_reason?: string;
  reversed_by?: string;
  source: string;
  created_by?: string;
  created_at: string;
  updated_at: string;
  lines?: JournalLine[];
}

export interface CreateJournalLine {
  account_id: string;
  debit: number;
  credit: number;
}

export interface CreateJournalRequest {
  date: string;
  description: string;
  invoice_id?: string;
  status?: string;
  lines: CreateJournalLine[];
}

export interface UpdateJournalRequest {
  date?: string;
  description?: string;
  lines?: CreateJournalLine[];
}

export interface ReverseJournalRequest {
  reason: string;
}

// ---- Customer ----

export interface Customer {
  id: string;
  name: string;
  email?: string;
  phone?: string;
  address?: string;
  is_active: boolean;
  created_by?: string;
  created_at: string;
  updated_at: string;
}

// ---- Invoice ----

export type InvoiceStatus = "unpaid" | "partially_paid" | "paid";

export interface Invoice {
  id: string;
  invoice_number: string;
  customer_id: string;
  customer_name?: string;
  issue_date: string;
  due_date: string;
  total_amount: number;
  amount_paid: number;
  status: InvoiceStatus;
  description?: string;
  days_overdue?: number;
  created_by?: string;
  created_at: string;
  updated_at: string;
}

// ---- Payment ----

export interface PaymentAllocation {
  id: string;
  payment_id: string;
  invoice_id: string;
  invoice_number?: string;
  payment_number?: string;
  amount: number;
  created_at: string;
}

export interface Payment {
  id: string;
  payment_number: string;
  customer_id: string;
  customer_name?: string;
  payment_date: string;
  amount: number;
  deposit_to_account_id: string;
  deposit_account_code?: string;
  deposit_account_name?: string;
  journal_entry_id?: string;
  notes?: string;
  allocations?: PaymentAllocation[];
  created_by?: string;
  created_at: string;
  updated_at: string;
}

export interface PaymentAllocationLine {
  invoice_id: string;
  amount: number;
}

export interface CreatePaymentRequest {
  customer_id: string;
  payment_date: string;
  amount: number;
  deposit_to_account_id: string;
  notes?: string;
  allocations: PaymentAllocationLine[];
}

export interface ReceivablesSummary {
  total_outstanding: number;
  total_overdue: number;
}

// ---- Auth ----

export interface LoginRequest {
  username: string;
  password: string;
}

export interface AuthTokens {
  access_token: string;
  refresh_token: string;
}

// ---- Generic API response wrappers ----

export interface ApiResponse<T> {
  data: T;
  message?: string;
}

export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  limit: number;
  offset: number;
}

export interface ApiError {
  message: string;
  code?: string;
}
