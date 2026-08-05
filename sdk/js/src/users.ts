import { SaaSKitClient } from './client';

export interface User {
  id: string;
  email: string;
  first_name?: string;
  last_name?: string;
  email_verified: boolean;
  status: string;
  created_at: string;
  updated_at: string;
  metadata_public?: Record<string, any>;
}

export interface UpdateUserRequest {
  first_name?: string;
  last_name?: string;
}

export interface Session {
  id: string;
  user_agent: string;
  ip_address: string;
  last_active_at: string;
  created_at: string;
  expires_at: string;
}

export class UsersClient {
  constructor(private client: SaaSKitClient) {}

  async getMe(accessToken: string): Promise<User> {
    const response = await this.client.request<User>(
      '/api/v1/users/me',
      {},
      accessToken
    );
    return response.data;
  }

  async updateMe(
    accessToken: string,
    request: UpdateUserRequest
  ): Promise<User> {
    const response = await this.client.request<User>(
      '/api/v1/users/me',
      {
        method: 'PATCH',
        body: JSON.stringify(request),
      },
      accessToken
    );
    return response.data;
  }

  async listSessions(accessToken: string): Promise<Session[]> {
    const response = await this.client.request<{ sessions: Session[] }>(
      '/api/v1/users/me/sessions',
      {},
      accessToken
    );
    return response.data.sessions;
  }

  async revokeSession(
    accessToken: string,
    sessionId: string
  ): Promise<void> {
    await this.client.request<void>(
      `/api/v1/users/me/sessions/${sessionId}`,
      {
        method: 'DELETE',
      },
      accessToken
    );
  }
}
