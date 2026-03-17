// ---------------------------------------------------------------------------
// Invoices API service
// ---------------------------------------------------------------------------

import { apiGet, apiPost } from "../api-client";
import type { Invoice } from "../types";

interface InvoiceListParams {
  customer_id?: string;
  status?: string;
  limit?: string;
  offset?: string;
}

export interface CreateInvoiceRequest {
  customer_id: string;
  issue_date: string;   // YYYY-MM-DD
  due_date: string;     // YYYY-MM-DD
  total_amount: number;
  description?: string;
}

export const invoicesApi = {
  list(params?: InvoiceListParams) {
    return apiGet<Invoice[]>("/invoices", params as Record<string, string>);
  },

  getById(id: string) {
    return apiGet<Invoice>(`/invoices/${id}`);
  },

  create(data: CreateInvoiceRequest) {
    return apiPost<Invoice>("/invoices", data);
  },
};
