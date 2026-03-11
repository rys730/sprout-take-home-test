// ---------------------------------------------------------------------------
// Auth API service
// ---------------------------------------------------------------------------

import { apiPost } from "../api-client";
import type { AuthTokens, LoginRequest } from "../types";

export const authApi = {
  login(data: LoginRequest) {
    return apiPost<AuthTokens>("/auth/login", data);
  },

  refreshToken(refreshToken: string) {
    return apiPost<AuthTokens>("/auth/refresh", { refresh_token: refreshToken });
  },
};
