import { SaaSKitClient } from './client';

export interface UserMetadata {
  metadata_public?: Record<string, any>;
  metadata_private?: Record<string, any>;
}

export interface UpdateUserMetadataRequest {
  metadata_public?: Record<string, any>;
  metadata_private?: Record<string, any>;
}

export interface TenantMetadata {
  metadata?: Record<string, any>;
}

export interface UpdateTenantMetadataRequest {
  metadata?: Record<string, any>;
}

export class MetadataClient {
  constructor(private client: SaaSKitClient) {}

  async getUserMetadata(accessToken: string): Promise<UserMetadata> {
    const response = await this.client.request<UserMetadata>(
      '/api/v1/users/me/metadata',
      {},
      accessToken
    );
    return response.data;
  }

  async updateUserMetadata(
    accessToken: string,
    request: UpdateUserMetadataRequest
  ): Promise<UserMetadata> {
    const response = await this.client.request<UserMetadata>(
      '/api/v1/users/me/metadata',
      {
        method: 'PATCH',
        body: JSON.stringify(request),
      },
      accessToken
    );
    return response.data;
  }

  async getTenantMetadata(
    accessToken: string,
    tenantId: string
  ): Promise<TenantMetadata> {
    const response = await this.client.request<TenantMetadata>(
      `/api/v1/tenants/${tenantId}/metadata`,
      {},
      accessToken
    );
    return response.data;
  }

  async updateTenantMetadata(
    accessToken: string,
    tenantId: string,
    request: UpdateTenantMetadataRequest
  ): Promise<TenantMetadata> {
    const response = await this.client.request<TenantMetadata>(
      `/api/v1/tenants/${tenantId}/metadata`,
      {
        method: 'PATCH',
        body: JSON.stringify(request),
      },
      accessToken
    );
    return response.data;
  }
}
