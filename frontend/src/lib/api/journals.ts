// ---------------------------------------------------------------------------
// Journals API service
// ---------------------------------------------------------------------------

import { apiGet, apiPost, apiPut, apiDelete } from "../api-client";
import type {
  JournalEntry,
  CreateJournalRequest,
  UpdateJournalRequest,
  ReverseJournalRequest,
} from "../types";

interface JournalListParams {
  status?: string;
  source?: string;
  start_date?: string;
  end_date?: string;
  limit?: string;
  offset?: string;
}

export const journalsApi = {
  list(params?: JournalListParams) {
    return apiGet<JournalEntry[]>("/journals", params as Record<string, string>);
  },

  getById(id: string) {
    return apiGet<JournalEntry>(`/journals/${id}`);
  },

  create(data: CreateJournalRequest) {
    return apiPost<JournalEntry>("/journals", data);
  },

  update(id: string, data: UpdateJournalRequest) {
    return apiPut<JournalEntry>(`/journals/${id}`, data);
  },

  post(id: string) {
    return apiPost<JournalEntry>(`/journals/${id}/post`);
  },

  reverse(id: string, data: ReverseJournalRequest) {
    return apiPost<JournalEntry>(`/journals/${id}/reverse`, data);
  },

  delete(id: string) {
    return apiDelete(`/journals/${id}`);
  },
};
