// ---------------------------------------------------------------------------
// Accounts API service
// ---------------------------------------------------------------------------

import { apiGet, apiPost, apiPut, apiDelete } from "../api-client";
import type {
  Account,
  AccountTreeNode,
  CreateAccountRequest,
  UpdateAccountRequest,
} from "../types";

export const accountsApi = {
  list(params?: { search?: string; type?: string }) {
    return apiGet<Account[]>("/accounts", params as Record<string, string>);
  },

  getTree() {
    return apiGet<AccountTreeNode[]>("/accounts/tree");
  },

  getById(id: string) {
    return apiGet<Account>(`/accounts/${id}`);
  },

  create(data: CreateAccountRequest) {
    return apiPost<Account>("/accounts", data);
  },

  update(id: string, data: UpdateAccountRequest) {
    return apiPut<Account>(`/accounts/${id}`, data);
  },

  delete(id: string) {
    return apiDelete(`/accounts/${id}`);
  },
};
