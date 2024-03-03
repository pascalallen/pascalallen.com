import { jwtDecode } from 'jwt-decode';
import moment from 'moment';
import { listToMap } from '@utilities/collections';
import request from '@utilities/request';
import HttpMethod from '@domain/constants/HttpMethod';
import { User } from '@domain/types/User';
import AuthStore from '@stores/AuthStore';
import { LoginFormValues } from '@pages/LoginPage';

// TODO: Extract to reset password component
export type ResetPasswordFormValues = {
  token: string;
  password: string;
  confirm_password: string;
};

export type AuthenticatedResponsePayload = {
  access_token: string;
  refresh_token: string;
  user: User;
  roles?: string[];
  permissions?: string[];
};

class AuthService {
  private readonly authStore: AuthStore;

  constructor(authStore: AuthStore) {
    this.authStore = authStore;
  }

  public async login(params: LoginFormValues): Promise<void> {
    const response = await request.send<AuthenticatedResponsePayload>({
      method: HttpMethod.POST,
      uri: '/api/v1/auth/login',
      body: params,
      options: { auth: false }
    });
    this.authStore.setData({
      access_token: response.body.data?.access_token ?? '',
      refresh_token: response.body.data?.refresh_token ?? '',
      access_token_claims: jwtDecode(response.body.data?.access_token ?? ''),
      client_iat: moment().unix(),
      roles: listToMap(response.body.data?.roles ?? []),
      permissions: listToMap(response.body.data?.permissions ?? [])
    });
  }

  public async logout(): Promise<void> {
    this.authStore.clearData();
  }

  public async refresh(): Promise<void> {
    const response = await request.send<AuthenticatedResponsePayload>({
      method: HttpMethod.PATCH,
      uri: '/api/v1/auth/refresh',
      body: {
        access_token: this.authStore.getData()?.access_token ?? '',
        refresh_token: this.authStore.getData()?.refresh_token ?? ''
      },
      options: { auth: false }
    });
    this.authStore.setData({
      access_token: response.body.data?.access_token ?? '',
      refresh_token: response.body.data?.refresh_token ?? '',
      access_token_claims: jwtDecode(response.body.data?.access_token ?? ''),
      client_iat: moment().unix(),
      roles: listToMap(response.body.data?.roles ?? []),
      permissions: listToMap(response.body.data?.permissions ?? [])
    });
  }

  public async requestReset(emailAddress: string): Promise<void> {
    await request.send<void>({
      method: HttpMethod.POST,
      uri: '/api/v1/auth/request-reset',
      body: {
        email_address: emailAddress
      },
      options: { auth: false }
    });
  }

  public async resetPassword(params: ResetPasswordFormValues): Promise<void> {
    const response = await request.send<AuthenticatedResponsePayload>({
      method: HttpMethod.POST,
      uri: '/api/v1/auth/reset-password',
      body: params,
      options: { auth: false }
    });
    this.authStore.setData({
      access_token: response.body.data?.access_token ?? '',
      refresh_token: response.body.data?.refresh_token ?? '',
      access_token_claims: jwtDecode(response.body.data?.access_token ?? ''),
      client_iat: moment().unix(),
      roles: listToMap(response.body.data?.roles ?? []),
      permissions: listToMap(response.body.data?.permissions ?? [])
    });
  }

  public isLoggedIn(): boolean {
    return this.authStore.hasData();
  }

  public hasPermission(permission: string): boolean {
    if (!this.authStore.hasData()) {
      return false;
    }

    return this.authStore.getData()?.permissions[permission] === permission;
  }

  public hasPermissions(permissions: string[]): boolean {
    if (!this.authStore.hasData()) {
      return false;
    }

    return permissions.every((permission: string) => {
      return this.authStore.getData()?.permissions[permission] === permission;
    });
  }

  public hasRole(role: string): boolean {
    if (!this.authStore.hasData()) {
      return false;
    }

    return this.authStore.getData()?.roles[role] === role;
  }
}

export default AuthService;
