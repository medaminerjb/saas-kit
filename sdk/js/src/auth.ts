import { SaaSKitClient, SaaSKitError } from './client';

export interface RegisterRequest {
  email: string;
  password: string;
}

export interface RegisterResponse {
  user_id: string;
  email: string;
  message: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface LoginResponse {
  access_token: string;
  refresh_token: string;
  token_type: string;
  expires_in: number;
}

export interface RefreshRequest {
  refresh_token: string;
}

export interface RefreshResponse {
  access_token: string;
  refresh_token: string;
  token_type: string;
  expires_in: number;
}

export interface LogoutRequest {
  refresh_token: string;
}

export interface ForgotPasswordRequest {
  email: string;
}

export interface ResetPasswordRequest {
  token: string;
  password: string;
}

export interface VerifyEmailRequest {
  token: string;
}

export class AuthClient {
  constructor(private client: SaaSKitClient) {}

  async register(request: RegisterRequest): Promise<RegisterResponse> {
    const response = await this.client.request<RegisterResponse>(
      '/api/v1/auth/register',
      {
        method: 'POST',
        body: JSON.stringify(request),
      }
    );
    return response.data;
  }

  async login(request: LoginRequest): Promise<LoginResponse> {
    const response = await this.client.request<LoginResponse>(
      '/api/v1/auth/login',
      {
        method: 'POST',
        body: JSON.stringify(request),
      }
    );
    return response.data;
  }

  async refresh(request: RefreshRequest): Promise<RefreshResponse> {
    const response = await this.client.request<RefreshResponse>(
      '/api/v1/auth/refresh',
      {
        method: 'POST',
        body: JSON.stringify(request),
      }
    );
    return response.data;
  }

  async logout(
    accessToken: string,
    request: LogoutRequest
  ): Promise<void> {
    await this.client.request<void>(
      '/api/v1/auth/logout',
      {
        method: 'POST',
        body: JSON.stringify(request),
      },
      accessToken
    );
  }

  async forgotPassword(request: ForgotPasswordRequest): Promise<void> {
    await this.client.request<void>(
      '/api/v1/auth/forgot-password',
      {
        method: 'POST',
        body: JSON.stringify(request),
      }
    );
  }

  async resetPassword(request: ResetPasswordRequest): Promise<void> {
    await this.client.request<void>(
      '/api/v1/auth/reset-password',
      {
        method: 'POST',
        body: JSON.stringify(request),
      }
    );
  }

  async verifyEmail(request: VerifyEmailRequest): Promise<void> {
    await this.client.request<void>(
      '/api/v1/auth/verify-email',
      {
        method: 'POST',
        body: JSON.stringify(request),
      }
    );
  }
}
