// ---------------------------------------------------------------------------
// Invoices API service
// ---------------------------------------------------------------------------

import { apiGet, apiPost } from "../api-client";
import type { Invoice } from "../types";

interface InvoiceListParams {
  customer_id?: string;
  status?: string | string[];
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
    const { status, ...rest } = params ?? {};
    const flat: Record<string, string> = { ...rest } as Record<string, string>;
    if (status) {
      flat.status = Array.isArray(status) ? status.join(",") : status;
    }
    return apiGet<Invoice[]>("/invoices", flat);
  },

  getById(id: string) {
    return apiGet<Invoice>(`/invoices/${id}`);
  },

  create(data: CreateInvoiceRequest) {
    return apiPost<Invoice>("/invoices", data);
  },
};
