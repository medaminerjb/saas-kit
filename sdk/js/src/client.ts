export interface SaaSKitConfig {
  baseURL?: string;
  apiKey?: string;
  timeout?: number;
  maxRetries?: number;
  retryDelay?: number;
}

export interface SaaSKitResponse<T> {
  data: T;
  status: number;
  headers: Headers;
}

export class SaaSKitError extends Error {
  constructor(
    public message: string,
    public status?: number,
    public response?: any
  ) {
    super(message);
    this.name = 'SaaSKitError';
  }
}

export class SaaSKitClient {
  private config: Required<SaaSKitConfig>;
  private baseURL: string;

  public auth: AuthClient;
  public users: UsersClient;
  public tenants: TenantsClient;
  public metadata: MetadataClient;

  constructor(config: SaaSKitConfig = {}) {
    this.config = {
      baseURL: config.baseURL || 'http://localhost:8080',
      apiKey: config.apiKey || '',
      timeout: config.timeout || 30000,
      maxRetries: config.maxRetries || 3,
      retryDelay: config.retryDelay || 1000,
    };

    this.baseURL = this.config.baseURL;

    // Initialize API clients
    this.auth = new AuthClient(this);
    this.users = new UsersClient(this);
    this.tenants = new TenantsClient(this);
    this.metadata = new MetadataClient(this);
  }

  getBaseURL(): string {
    return this.baseURL;
  }

  getConfig(): Required<SaaSKitConfig> {
    return this.config;
  }

  async request<T>(
    endpoint: string,
    options: RequestInit = {},
    accessToken?: string
  ): Promise<SaaSKitResponse<T>> {
    const url = `${this.baseURL}${endpoint}`;
    const headers: HeadersInit = {
      'Content-Type': 'application/json',
      ...options.headers,
    };

    if (accessToken) {
      (headers as Record<string, string>)['Authorization'] = `Bearer ${accessToken}`;
    }

    if (this.config.apiKey) {
      (headers as Record<string, string>)['X-API-Key'] = this.config.apiKey;
    }

    let lastError: Error | null = null;

    for (let attempt = 0; attempt <= this.config.maxRetries; attempt++) {
      if (attempt > 0) {
        await this.backoff(attempt);
      }

      try {
        const response = await fetch(url, {
          ...options,
          headers,
          signal: AbortSignal.timeout(this.config.timeout),
        });

        // Don't retry on success or client errors (4xx)
        if (response.status < 500) {
          const data = await response.json();
          
          if (!response.ok) {
            throw new SaaSKitError(
              data.message || 'Request failed',
              response.status,
              data
            );
          }

          return {
            data,
            status: response.status,
            headers: response.headers,
          };
        }

        lastError = new SaaSKitError(`HTTP ${response.status}`, response.status);
      } catch (error) {
        lastError = error as Error;
        if (error instanceof SaaSKitError && error.status && error.status < 500) {
          throw error;
        }
      }
    }

    throw lastError || new SaaSKitError('Request failed after retries');
  }

  private backoff(attempt: number): Promise<void> {
    // Exponential backoff: delay * 2^(attempt-1)
    const delay = Math.min(
      this.config.retryDelay * Math.pow(2, attempt - 1),
      30000 // Cap at 30 seconds
    );
    return new Promise(resolve => setTimeout(resolve, delay));
  }
}

// Import other clients
import { AuthClient } from './auth';
import { UsersClient } from './users';
import { TenantsClient } from './tenants';
import { MetadataClient } from './metadata';
