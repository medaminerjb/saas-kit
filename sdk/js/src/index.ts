export { SaaSKitClient, SaaSKitConfig, SaaSKitResponse, SaaSKitError } from './client';
export { AuthClient, RegisterRequest, RegisterResponse, LoginRequest, LoginResponse, RefreshRequest, RefreshResponse, LogoutRequest, ForgotPasswordRequest, ResetPasswordRequest, VerifyEmailRequest } from './auth';
export { UsersClient, User, UpdateUserRequest, Session } from './users';
export { TenantsClient, Tenant, TenantWithRole, CreateTenantRequest, UpdateTenantRequest, SwitchTenantRequest, AcceptInvitationRequest, Member, InviteMemberRequest } from './tenants';
export { MetadataClient, UserMetadata, UpdateUserMetadataRequest, TenantMetadata, UpdateTenantMetadataRequest } from './metadata';

export default SaaSKitClient;
