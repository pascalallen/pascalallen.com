import {AccessTokenClaims} from "@domain/types/AccessTokenClaims";
import {RefreshTokenClaims} from "@domain/types/RefreshTokenClaims";

export type AuthData = {
    access_token: string;
    refresh_token: string;
    access_token_claims: AccessTokenClaims;
    refresh_token_claims: RefreshTokenClaims; // TODO: Revaluate whether this is necessary
    client_iat: number;
    roles: { [name: string]: string };
    permissions: { [name: string]: string };
};
