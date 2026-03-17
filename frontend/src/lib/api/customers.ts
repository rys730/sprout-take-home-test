// ---------------------------------------------------------------------------
// Customers API service
// ---------------------------------------------------------------------------

import { apiGet, apiPost, apiPut, apiDelete } from "../api-client";
import type { Customer } from "../types";

interface CustomerListParams {
  search?: string;
  limit?: string;
  offset?: string;
}

export const customersApi = {
  list(params?: CustomerListParams) {
    return apiGet<Customer[]>("/customers", params as Record<string, string>);
  },

  getById(id: string) {
    return apiGet<Customer>(`/customers/${id}`);
  },

  create(data: { name: string; email?: string; phone?: string; address?: string }) {
    return apiPost<Customer>("/customers", data);
  },

  update(id: string, data: { name?: string; email?: string; phone?: string; address?: string }) {
    return apiPut<Customer>(`/customers/${id}`, data);
  },

  delete(id: string) {
    return apiDelete(`/customers/${id}`);
  },
};
