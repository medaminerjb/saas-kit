import { SaaSKitClient } from './client';

export interface Tenant {
  id: string;
  name: string;
  slug: string;
  status: string;
  created_at: string;
  updated_at: string;
  metadata?: Record<string, any>;
}

export interface TenantWithRole {
  tenant: Tenant;
  role: string;
}

export interface CreateTenantRequest {
  name: string;
  slug?: string;
}

export interface UpdateTenantRequest {
  name?: string;
  slug?: string;
}

export interface SwitchTenantRequest {
  tenant_id: string;
}

export interface AcceptInvitationRequest {
  token: string;
}

export interface Member {
  id: string;
  tenant_id: string;
  user_id: string;
  email: string;
  role: string;
  created_at: string;
  updated_at: string;
}

export interface InviteMemberRequest {
  email: string;
  role: string;
}

export class TenantsClient {
  constructor(private client: SaaSKitClient) {}

  async create(
    accessToken: string,
    request: CreateTenantRequest
  ): Promise<Tenant> {
    const response = await this.client.request<Tenant>(
      '/api/v1/tenants',
      {
        method: 'POST',
        body: JSON.stringify(request),
      },
      accessToken
    );
    return response.data;
  }

  async list(accessToken: string): Promise<TenantWithRole[]> {
    const response = await this.client.request<{ tenants: TenantWithRole[] }>(
      '/api/v1/tenants',
      {},
      accessToken
    );
    return response.data.tenants;
  }

  async get(accessToken: string, tenantId: string): Promise<Tenant> {
    const response = await this.client.request<Tenant>(
      `/api/v1/tenants/${tenantId}`,
      {},
      accessToken
    );
    return response.data;
  }

  async update(
    accessToken: string,
    tenantId: string,
    request: UpdateTenantRequest
  ): Promise<Tenant> {
    const response = await this.client.request<Tenant>(
      `/api/v1/tenants/${tenantId}`,
      {
        method: 'PATCH',
        body: JSON.stringify(request),
      },
      accessToken
    );
    return response.data;
  }

  async switch(
    accessToken: string,
    request: SwitchTenantRequest
  ): Promise<void> {
    await this.client.request<void>(
      '/api/v1/tenants/switch',
      {
        method: 'POST',
        body: JSON.stringify(request),
      },
      accessToken
    );
  }

  async acceptInvitation(
    accessToken: string,
    request: AcceptInvitationRequest
  ): Promise<void> {
    await this.client.request<void>(
      '/api/v1/tenants/invitations/accept',
      {
        method: 'POST',
        body: JSON.stringify(request),
      },
      accessToken
    );
  }

  async listMembers(accessToken: string, tenantId: string): Promise<Member[]> {
    const response = await this.client.request<{ members: Member[] }>(
      `/api/v1/tenants/${tenantId}/members`,
      {},
      accessToken
    );
    return response.data.members;
  }

  async inviteMember(
    accessToken: string,
    tenantId: string,
    request: InviteMemberRequest
  ): Promise<Member> {
    const response = await this.client.request<Member>(
      `/api/v1/tenants/${tenantId}/members`,
      {
        method: 'POST',
        body: JSON.stringify(request),
      },
      accessToken
    );
    return response.data;
  }

  async removeMember(
    accessToken: string,
    tenantId: string,
    userId: string
  ): Promise<void> {
    await this.client.request<void>(
      `/api/v1/tenants/${tenantId}/members/${userId}`,
      {
        method: 'DELETE',
      },
      accessToken
    );
  }
}
