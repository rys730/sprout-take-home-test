// ---------------------------------------------------------------------------
// Payments API service
// ---------------------------------------------------------------------------

import { apiGet, apiPost } from "../api-client";
import type {
  Payment,
  CreatePaymentRequest,
  ReceivablesSummary,
} from "../types";

interface PaymentListParams {
  customer_id?: string;
  limit?: string;
  offset?: string;
}

export const paymentsApi = {
  list(params?: PaymentListParams) {
    return apiGet<Payment[]>("/payments", params as Record<string, string>);
  },

  getById(id: string) {
    return apiGet<Payment>(`/payments/${id}`);
  },

  record(data: CreatePaymentRequest) {
    return apiPost<Payment>("/payments", data);
  },

  getSummary() {
    return apiGet<ReceivablesSummary>("/payments/summary");
  },
};
