import { AccessTokenClaims } from './AccessTokenClaims';

export type AuthData = {
  access_token: string;
  refresh_token: string;
  access_token_claims: AccessTokenClaims;
  client_iat: number;
  roles: { [name: string]: string };
  permissions: { [name: string]: string };
};
